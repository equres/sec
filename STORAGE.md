# File System Options

We have a very large number of files that we need to keep in order to properly serve the necessary data for our website. 

The main files we have in our system are the following:
- **Index Files**: These files include the main data, such as the filings, the companies, and the file links.
- **Filing/Data Files**: These files include the actual data, such as the company details.
- **ZIP Files**: These files are compressed ZIP files that include the files for that specific filing.

In our case, we need to mainly have the Index files as well as the Filing/Data Files. So need to figure our how to shape your file system.

_Note: We will assume that Index files will be downloaded and will exist regardless of the options we have because we need those files in  order to retrieve any data._

## Option 1
The first option, which is what we have right now, is to focus on downloading ZIP files and extracting them, then whatever file we are missing from there, we download them individually, such as in `sec dow data`.

Advantages:
- We can reduce the amount of time we need to wait in order to download the data, that is because we download a compressed version of, let us say, 6 files, instead of downloading them individually.
- If we use `rsync` or mirroring in order to backup our data to a different server, then it would be faster if we move the ZIP files, and then extracting them again on the new server.

Disadvantages:
- We cannot properly compress the files because there are already ZIP files in the directory.


## Option 2
The second option is to download the files individually without downloading any ZIP files. This can be done by running `sec dow data`.

Advantages:
- The application would be simpler, since we do not have to worry about downloading the ZIP files and making sure they have the correct files.
- It would be easier to backup the data, since we can simply compress the directory and move it to a different server.

Disadvantages:
- The download process would take much longer since we would have to download the files individually.

## Option 3
This option can be used along with either of these two options mentioned above. It involves using ZFS, which is a more sophisticated file system that uses things such a snapshots and data integrity checks.

We would divide our data into three parts:
- SEC Files
- Database
- Backups

And inside the backups, we would have snapshots of the ZFS file system that we can use to restore the data. We can also download and set up a package called `zfs-auto-snapshot` to automate that for us.