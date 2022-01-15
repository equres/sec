ALTER TABLE sec.secItemFile 
ADD CONSTRAINT fk_cik FOREIGN KEY (cikNumber) REFERENCES sec.ciks(cik);