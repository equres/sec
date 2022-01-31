package secdata

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"

	"github.com/equres/sec/pkg/sec"
	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
)

type SECDataOpsFSDS struct {
	DataType string
}

func NewSECDataOpsFSDS() *SECDataOpsFSDS {
	return &SECDataOpsFSDS{
		DataType: "fsds",
	}
}

func (s *SECDataOpsFSDS) GetDataType() string {
	return s.DataType
}

func (s *SECDataOpsFSDS) GetDataFilePath(baseURL string, yearQuarter string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("/files/dera/data/financial-statement-data-sets/%v.zip", yearQuarter)

	pathURL, err := url.Parse(filePath)
	if err != nil {
		return "", err
	}
	return parsedURL.ResolveReference(pathURL).String(), nil
}

func (s *SECDataOpsFSDS) GetDataDirPath() string {
	return "files/dera/data/financial-statement-data-sets/"
}

func (s *SECDataOpsFSDS) GetDataTypeInsertFunc(fileName string) func(*sec.SEC, *sqlx.DB, io.ReadCloser) error {
	switch fileName {
	case "sub.txt":
		return FSDSSubDataUpsert
	case "tag.txt":
		return FSDSTagDataUpsert
	case "num.txt":
		return FSDSNumDataUpsert
	case "pre.txt":
		return FSDSPreDataUpsert
	default:
		return nil
	}
}

func FSDSSubDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	subs := []FSDSSUB{}
	err := gocsv.Unmarshal(reader, &subs)
	if err != nil {
		return err
	}

	for _, v := range subs {
		var period sql.NullString
		if v.Period != "" {
			period = sql.NullString{
				String: v.Period,
			}
		}

		var filled sql.NullString
		if v.Filled != "" {
			filled = sql.NullString{
				String: v.Filled,
			}
		}

		_, err = db.Exec(`
		INSERT INTO fsds.sub (adsh, cik, name, sic, countryba, stprba, cityba, zipba, bas1, bas2, baph, countryma, strpma, cityma, zipma, mas1, mas2, countryinc, stprinc, ein, former, changed, afs, wksi, fye, form, period, fy, fp, filled, accepted, prevrpt, detail, instance, nciks, aciks, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, NOW(), NOW()) 
		ON CONFLICT (adsh, cik, name, sic) 
		DO NOTHING;`, v.Adsh, v.CIK, v.Name, v.SIC, v.CountryBA, v.StprBA, v.CityBA, v.ZIPBA, v.BAS1, v.BAS2, v.BAPH, v.CountryMA, v.StrpMA, v.CityMA, v.ZIPMA, v.MAS1, v.MAS2, v.CountryInc, v.StprInc, v.EIN, v.Former, v.Changed, v.Afs, v.Wksi, v.Fye, v.Form, period, v.Fy, v.Fp, filled, v.Accepted, v.Prevrpt, v.Detail, v.Instance, v.Nciks, v.Aciks)
		if err != nil {
			return err
		}
	}
	return nil
}

func FSDSTagDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	tags := []FSDSTAG{}
	err := gocsv.Unmarshal(reader, &tags)
	if err != nil {
		return err
	}

	for _, v := range tags {
		_, err = db.Exec(`
			INSERT INTO fsds.tag (tag, version, custom, abstract, datatype, lord, crdr, tlabel, doc, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (tag, version) 
			DO NOTHING;`, v.Tag, v.Version, v.Custom, v.Abstract, v.Datatype, v.Lord, v.Crdr, v.Tlabel, v.Doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func FSDSNumDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	nums := []FSDSNUM{}
	err := gocsv.Unmarshal(reader, &nums)
	if err != nil {
		return err
	}

	for _, v := range nums {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM fsds.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO fsds.num (adsh, tag, version, coreg, ddate, qtrs, uom, value, footnote, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, tag, version, coreg, ddate, qtrs, uom) 
			DO NOTHING;`, v.Adsh, v.Tag, v.Version, v.Coreg, v.DDate, v.Qtrs, v.UOM, v.Value, v.Footnote)
		if err != nil {
			return err
		}
	}
	return nil
}

func FSDSPreDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	pres := []FSDSPRE{}
	err := gocsv.Unmarshal(reader, &pres)
	if err != nil {
		return err
	}

	for _, v := range pres {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM fsds.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO fsds.pre (adsh, report, line, stmt, inpth, rfile, tag, version, plabel, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, report, line) 
			DO NOTHING;`, v.Adsh, v.Report, v.Line, v.Stmt, v.Inpth, v.Rfile, v.Tag, v.Version, v.Plabel)
		if err != nil {
			return err
		}
	}
	return nil
}
