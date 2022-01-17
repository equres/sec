package secextra

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secevent"
	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type SUB struct {
	Adsh       string `csv:"adsh"`
	CIK        int    `csv:"cik"`
	Name       string `csv:"name"`
	SIC        string `csv:"sic"`
	CountryBA  string `csv:"countryba"`
	StprBA     string `csv:"stprba"`
	CityBA     string `csv:"cityba"`
	ZIPBA      string `csv:"zipba"`
	BAS1       string `csv:"bas1"`
	BAS2       string `csv:"bas2"`
	BAPH       string `csv:"baph"`
	CountryMA  string `csv:"countryma"`
	StrpMA     string `csv:"strpma"`
	CityMA     string `csv:"cityma"`
	ZIPMA      string `csv:"zipma"`
	MAS1       string `csv:"mas1"`
	MAS2       string `csv:"mas2"`
	CountryInc string `csv:"countryinc"`
	StprInc    string `csv:"stprinc"`
	EIN        string `csv:"ein"`
	Former     string `csv:"former"`
	Changed    string `csv:"changed"`
	Afs        string `csv:"afs"`
	Wksi       string `csv:"wksi"`
	Fye        string `csv:"fye"`
	Form       string `csv:"form"`
	Period     string `csv:"period"`
	Fy         string `csv:"fy"`
	Fp         string `csv:"fp"`
	Filled     string `csv:"filled"`
	Accepted   string `csv:"accepted"`
	Prevrpt    string `csv:"prevrpt"`
	Detail     string `csv:"detail"`
	Instance   string `csv:"instance"`
	Nciks      string `csv:"nciks"`
	Aciks      string `csv:"aciks"`
}

type NUM struct {
	Adsh     string `csv:"adsh"`
	Tag      string `csv:"tag"`
	Version  string `csv:"version"`
	Coreg    string `csv:"coreg"`
	DDate    string `csv:"ddate"`
	Qtrs     string `csv:"qtrs"`
	UOM      string `csv:"uom"`
	Value    string `csv:"value"`
	Footnote string `csv:"footnote"`
}

type TAG struct {
	Tag      string `csv:"tag"`
	Version  string `csv:"version"`
	Custom   string `csv:"custom"`
	Abstract string `csv:"abstract"`
	Datatype string `csv:"datatype"`
	Lord     string `csv:"lord"`
	Crdr     string `csv:"crdr"`
	Tlabel   string `csv:"tlabel"`
	Doc      string `csv:"doc"`
}

type PRE struct {
	Adsh    string `csv:"adsh"`
	Report  string `csv:"report"`
	Line    string `csv:"line"`
	Stmt    string `csv:"stmt"`
	Inpth   string `csv:"inpth"`
	Rfile   string `csv:"rfile"`
	Tag     string `csv:"tag"`
	Version string `csv:"version"`
	Plabel  string `csv:"plabel"`
}

func IndexFinancialStatementDataSets(s *sec.SEC, db *sqlx.DB) error {
	filesPath := filepath.Join(s.Config.Main.CacheDir, "files/dera/data/financial-statement-data-sets/")
	files, err := ioutil.ReadDir(filesPath)
	if err != nil {
		return err
	}
	for _, v := range files {
		if s.Verbose {
			log.Info(fmt.Sprintf("Indexing file %v: ", v.Name()))
		}
		reader, err := zip.OpenReader(filepath.Join(filesPath, v.Name()))
		if err != nil {
			return err
		}

		err = FinancialStatementDataSetsZIPUpsert(s, db, filesPath, reader.File)
		if err != nil {
			secevent.CreateIndexEvent(db, filesPath, "failed", "error_inserting_financial_statements_in_database")
			return err
		}

		reader.Close()

		if s.Verbose {
			log.Info("\u2713")
		}
		secevent.CreateIndexEvent(db, filesPath, "success", "")
	}
	return nil
}

func FinancialStatementDataSetsZIPUpsert(s *sec.SEC, db *sqlx.DB, pathname string, files []*zip.File) error {
	for _, file := range files {
		fileName := strings.ToLower(file.Name)

		if fileName == "readme.htm" || fileName == "readme.html" {
			continue
		}

		var upsertFunc func(*sec.SEC, *sqlx.DB, io.ReadCloser) error
		switch fileName {
		case "sub.txt":
			upsertFunc = SubDataUpsert
		case "tag.txt":
			upsertFunc = TagDataUpsert
		case "num.txt":
			upsertFunc = NumDataUpsert
		case "pre.txt":
			upsertFunc = PreDataUpsert
		default:
			continue
		}

		if s.Verbose {
			log.Info(fmt.Sprintf("Indexing file %v\n", fileName))
		}

		reader, err := file.Open()
		if err != nil {
			secevent.CreateIndexEvent(db, pathname, "failed", "could_not_open_zip_file")
			return err
		}

		err = upsertFunc(s, db, reader)
		if err != nil {
			return err
		}

		reader.Close()
	}
	secevent.CreateIndexEvent(db, pathname, "success", "")
	return nil
}

func SubDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	subs := []SUB{}
	err = gocsv.Unmarshal(reader, &subs)
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
		INSERT INTO sec.sub (adsh, cik, name, sic, countryba, stprba, cityba, zipba, bas1, bas2, baph, countryma, strpma, cityma, zipma, mas1, mas2, countryinc, stprinc, ein, former, changed, afs, wksi, fye, form, period, fy, fp, filled, accepted, prevrpt, detail, instance, nciks, aciks, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, NOW(), NOW()) 
		ON CONFLICT (adsh, cik, name, sic) 
		DO NOTHING;`, v.Adsh, v.CIK, v.Name, v.SIC, v.CountryBA, v.StprBA, v.CityBA, v.ZIPBA, v.BAS1, v.BAS2, v.BAPH, v.CountryMA, v.StrpMA, v.CityMA, v.ZIPMA, v.MAS1, v.MAS2, v.CountryInc, v.StprInc, v.EIN, v.Former, v.Changed, v.Afs, v.Wksi, v.Fye, v.Form, period, v.Fy, v.Fp, filled, v.Accepted, v.Prevrpt, v.Detail, v.Instance, v.Nciks, v.Aciks)
		if err != nil {
			return err
		}
	}
	return nil
}

func TagDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	tags := []TAG{}
	err = gocsv.Unmarshal(reader, &tags)
	if err != nil {
		return err
	}

	for _, v := range tags {
		_, err = db.Exec(`
			INSERT INTO sec.tag (tag, version, custom, abstract, datatype, lord, crdr, tlabel, doc, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (tag, version) 
			DO NOTHING;`, v.Tag, v.Version, v.Custom, v.Abstract, v.Datatype, v.Lord, v.Crdr, v.Tlabel, v.Doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func NumDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	nums := []NUM{}
	err = gocsv.Unmarshal(reader, &nums)
	if err != nil {
		return err
	}

	for _, v := range nums {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM sec.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO sec.num (adsh, tag, version, coreg, ddate, qtrs, uom, value, footnote, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, tag, version, coreg, ddate, qtrs, uom) 
			DO NOTHING;`, v.Adsh, v.Tag, v.Version, v.Coreg, v.DDate, v.Qtrs, v.UOM, v.Value, v.Footnote)
		if err != nil {
			return err
		}
	}
	return nil
}

func PreDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	pres := []PRE{}
	err = gocsv.Unmarshal(reader, &pres)
	if err != nil {
		return err
	}

	for _, v := range pres {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM sec.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO sec.pre (adsh, report, line, stmt, inpth, rfile, tag, version, plabel, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, report, line) 
			DO NOTHING;`, v.Adsh, v.Report, v.Line, v.Stmt, v.Inpth, v.Rfile, v.Tag, v.Version, v.Plabel)
		if err != nil {
			return err
		}
	}
	return nil
}
