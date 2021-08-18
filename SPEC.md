# SEC - SEC financial statement archive

`vim:set tw=1000:`

The purpose of the project is to create a simple platform with all SEC financial statements indexed,
and to offer a great searchability to the users of the platform. 

## Functional specs

Search functionality will be similar to this website:

https://www.sec.gov/search/search.htm

But be much more powerful, and also have more functionality, and look
simpler/better.

Upon searching, the number of matching filings will be shown, with a
searched terms highlighted, similar to Google (little paragraph of text
around the text that matched + a way to click the whole thing).
It's going to be possible to click it, and jump to the matching filing,
maybe in the place where the search string matched.

### Goal

Make a simple but powerful search engine for SEC stock filings.
The perfect interface to SECARCH would be a single text box that can accept 
a query like "avocado" just like Google, and use this query to search in all
statements.

Of course, more advanced queries could be allowed too: "avocado -grain
-wheat" would show companies that have something to do with avocado, but not
wheat or grain.

The special thing about SECARCH is that it'll be up-to-date with the
statements that are being published every day. So there must be a mechanism
for it to be fully automated, so that new statements are loaded into a
database.

### Non-goal:

- Pretty look - we will start with a simple brutalist look that is very fast to load
- JSON based API - I don't think we will ever need one for this stuff
- We will worry only about HTML/text reports. Don't worry about actually parsing XML

## Target

It's safe to assume for now we're talking about 1-user (me) website, no user accounts, no authentication,
plain CSS, minimal or no Javascript, Go templates and server-side generated HTML. No thrills on UI, but
with a data store that is very searchable.

Long term we could add accounts and payments and have people join to access our financial platform.

## Project milestones:

1. **Data synchronization pipeline (from SEC.gov to disk/backup)** (PRIORITY)
2. Data ingest pipeline (from disk/backup to PostgreSQL)
3. Web UI (from PostgreSQL to the user)

## Short intro

Companies fill financial filings and those are reported to the US
goverment agency called "SEC" (Securities and Exchange Commission) by filing
a report, and making it public.

The subsection of sec.gov that is responsible for reporting those results is called EDGAR.
If you see "EDGAR", it's their way of referring to a bunch of XML files available to the public.

The usual financial report talks about things that an investor may want to
know: what does the company do, what's their financial situations, how much
they borrowed, how much they sold, how much they spent etc.
All the filings are public, and available through SEC's website: https://sec.gov.

The format of the filings are:
- HTML (what is presented to the public -- this can be printed as a PDF and it's human readable)
- XML
- text

All the assets are included as well, so if there's a JPG included in the HTML website
(when for example the company wants to present a graph), this JPG will be there too.

The idea of the milestone 1 would be to understand what this RSS feed has,
and create a program to allow us to stay in sync with SEC by downloading the files mentioned
in the RSS feed.
It should be accomplished by parsing the feed of data and know what we need to download, and
then downloading it for a backup, and later for further ingesting.

The amount of data there is big, so we need to to maybe focus on 2021 data and get it going.
If it works for 2021 data, it should work for all past years.
In the case we could get a big server with lots of disk, and we could try it out downloading/indexing everything.



# Technical specs

## Data structure

The most important for this project is to understand the structure of the data.

Example of the RSS file is:

https://www.sec.gov/Archives/edgar/monthly/xbrlrss-2021-06.xml

Hierarchy:

- channel
  - item
    - `edgar:xbrlfiling`
      - `edgar:xbrlFiles`
        - `edgar:xbrlfile`
        - `edgar:xbrlfile`
        - `edgar:xbrlfile`
        - ...

The idea for a first version of the program would be to read the XML, parse the `edgar:xbrlfile` entries,
and download those files to the directory.

Example of those entries:

          <edgar:xbrlFile edgar:sequence="1" edgar:file="f10q0321a1_augmedixinc.htm" edgar:type="10-Q/A" edgar:size="461822" edgar:description="AMENDMENT NO. 1 TO FORM 10-Q" edgar:inlineXBRL="true" edgar:url="https://www.sec.gov/Archives/edgar/data/1769804/000121390021035115/f10q0321a1_augmedixinc.htm" />
          <edgar:xbrlFile edgar:sequence="2" edgar:file="f10q0321a1ex31-1_augmedix.htm" edgar:type="EX-31.1" edgar:size="9138" edgar:description="CERTIFICATION" edgar:url="https://www.sec.gov/Archives/edgar/data/1769804/000121390021035115/f10q0321a1ex31-1_augmedix.htm" />

We should grab those entries, and inject them to the DB. In other words, it'd be great if  `edgar:url` and download the files into `~/.sec/Archives/edgar.......`.

## Software architecture

The heart of `sec` will be a monolithic single-binary CLI called 'sec'.
The `sec` program will come with a couple of command line options which will
decide what the user is requesting.
We use Go and PostgreSQL as a heart of this program for managing its internal "index" of data.

PostgreSQL will hold:
- what stuff is in RSS feed
- what stuff we have already fetched
- what stuff we haven't yet fetched
- in the future (after milestone 1) it'll also hold all the data from ZIPs


Architecture diagram:

```

                                           PostgreSQL
                                              |
					      |
                                              4 
					      |
                                              |
sec.gov <---------2-----> SEC CLI Storage/backup <----3----> SQLite (for the index of all metadata)
  |
  1
  |
EDGAR
- ZIPs of:
  - xml
  - html
  - txt

```

The *top priority* is the 1+2+3. The 4 is not a priority now.

We assume we need to make the prototype of the pipeline working on 1 medium
size VPS server.


## Semantics

We should use Go libraries  `spf/cobra` and `spf/viper`:
- cobra will give us command line parsing required for the interface mentioned below
- viper will enable us a use of a config files


Proposed commands:

	sec init         # would create ~/.sec or ~/.config/sec/ and the config.yml there, with some default settings
	sec update       # get critical RSS feeds and download them ~/.sec/data directory, parse them and update the DB
	sec status       # show what we've got in the index with yyyy/mm 
	sec de yyyy/mm   # toggle "download enable" flag for statements from yyyy/mm month 
	sec dd yyyy/mm   # toggle "download disable" flag for statements from yyyy/mm month 
	sec dow          # download all statements with "download enable" flag
	
The example of something practical -- when we are going to be testing it on a VPS, we will do:

	sec init
	sec update
	sec dd all       # untoggle all years/months from downloading
	sec de 2020      # toggle for downloading all statements from 2020
	sec dest
	sec dow          # download everything
	
After 1-2 days, I'd like to be able to run:

	sec update
	
And it should spit out new entries in the RSS feed which we don't yet have. And then I should be able to do:

	sec dow

And it'd obtain this data locally.

## sec init

We should initialize a fresh system for using our program.
We will do this with `sec init`.
They should have management of "config" directories for the shell commands, and we should
ensure that `~/.sec/` or `~/.config/sec` can be created according to the OS system standards.
We could add config file there. Also `viper` I think has a way to dump config file with defaults
to disk.

We should for now assume PostgreSQL is installed on the same host, so it's all basically `localhost`
as a UNIX box, which will mimic the "local laptop" development environment.

The internal table for `sec init` could look ressemble this:

	CREATE TABLE sec.worklist (
		wid           SERIAL                 -- just an ID of the `worklist` row
		src_size      int                    -- orig size of the ZIP
		src_loc       TEXT NOT NULL          -- URL to the thing (ZIP?) that needs fetching/processing
		dst_loc       TEXT NOT NULL          -- path to where we write downloaded file
		
		
		
		
		will_download BOOL                   -- this is what "sec dd" or "sec de" will toggle
		
		created_at   TIMESTAMPTZ             -- when was this row created
		updated_at   TIMESTAMPTZ             -- when we run "sec update", we could maybe update this time
	)

`sec update` will grab the RSS feed, process the links to all the documents
that are mentioned there, and convert them to the rows in the `worklist`
table.
The worklist is basically a list of stuff we need to process.
I think the algo could be like:

	months_available = [EDGAR.start_date to current YYYY/MM]
	for each month and year in months_available:
		fetch RSS
		for each entry in this new downloaded RSS:
			is entry in sec.worklist?
				continue if so
			inject entry to sec.worklist with will_download = true
		
		
During the RSS processing, the good idea would be to use PostgresSQL's
UPSERT: if the row isn't there, we insert it.
If the line is there, we update it.

## sec dd / sec de

We should allow flipping `will_download` flag whenever we want to.

It should support 3 modes of toggling:

```
	sec dd all      # toggle all will_download off
	sec dd 2010.    # toggle just 1 year off
	sec dd 2012/12  # toggle 1 month off

```

The `dd` and `de` work the same way, but just the `will_download` value is different, so likely they should share the same code.

## sec dow

This is the most important command.
It would do:

	for each entry E in sec.worklist:
		U = submit HEAD request to  E.src_loc
		check if E.dst_loc is there && if E.size == U.size
			if so, it means sec.gov has the same file we've known about and we already have it downloaded
		download E.src_loc again and save it to E.dst_loc
		update E.size and toggle will_download = true		
		
In other words: this command will take a look at what's in `sec` and find unprocessed
items or items that are missing from a disk, and download them.
Rephrasing: we need to find inconsistencies between "us" and sec.gov and fix them.

# Resources

- https://www.sec.gov/os/accessing-edgar-data
- https://www.sec.gov/edgar/sec-api-documentation

## SKIP AFTER THIS LINE BECAUSE STUFF ISN'T YET READY ***

## Screens ***SKIP THIS FOR NOW***

There are couple of screens, but all of them are based on the similar looks
and feel.
First prototype won't have any Javascript.
Let's not focus on look at all.
All the HTML can be generated with Golang's 'template' package.

The idea is to have common navbar (top) and footer (bottom) and just have
the body to change

I'm making the resolution of screens in this spec be much smaller than a
real screen.
Don't worry about this--it's just to show the look of the side on
rectangular screens.
The final screens should have nothing to do with this.
Let's just simply display them without any scalling.
We may need some CSS, but it'll be provided after the first
prototype is written.

First screen: [screen_main.html](screen_main.html)

Results screen: [screen_results.html](screen_results.html)

## ## sec ingest

## sec server

The 'sec server' will start the webapp, and start listening on a default
port 8080.
It's possible to run 'sec server -listen :9000' to start listening on port
9000.

The interface to 'sec' is modelled after 'apt' command in Ubuntu Linux.
So 'sec update' will read the sec.gov website and figure out a list of things
which will need to be updated. The 'sec upgrade' will actually upgrade them
(meaning: it'll go and fetch the data from the sec.gov and load them to the
database).

The search should be based entirely on PostgreSQL Full-Text Search.
We can assume we're always going to have an access to PostgreSQL 12 or 13.
The data needs to


# Notes

- https://www.sedar.com/search/search_form_pc_en.htm -- Canadian SEC
