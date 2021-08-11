## Questions About Project/SPEC

1. What does disk/backup represent? Everytime we run the `sec update` we retrieve the data and save it directlt in the PostgreSQL database

2. The searching will be within the XML files right? For example https://www.sec.gov/Archives/edgar/monthly/xbrlrss-2021-07.xml 

3. How should we save the data (content that will be searched basically) for the XML files? For example we will have a list of rows in the DB:
    
    `Row 1: 2021 08 CONTENT`

    `Row 2: 2021 09 CONTENT`

4. If the search is done inside the XML files, then how will we direct the user where the searched word/phrase was found within the XML file if they clicked on it?

5. When I searched for the SEC of a company (Mister Car Wash) I did not find much useful information. I found an HTML page (https://www.sec.gov/Archives/edgar/data/1853513/000119312521178969/d117644ds1.htm) that showed details about the company and stuff. How will this file benefit us in the system (since it is not an XML file) or am I missing something (there is another file that perhaps I did not notice somehow)?

These are the questions I have for now, if I have more then will add here.
