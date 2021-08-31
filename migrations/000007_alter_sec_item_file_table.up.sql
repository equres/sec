ALTER TABLE sec.secItemFile
ADD CONSTRAINT xbrl_zip_file UNIQUE (cikNumber, accessionNumber, xbrlFile, xbrlSize);