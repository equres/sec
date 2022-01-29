package secdata

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secevent"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type MFDSUB struct {
	Adsh       string `csv:"adsh"`
	CIK        int    `csv:"cik"`
	Name       string `csv:"name"`
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
	Fye        string `csv:"fye"`
	PDate      string `csv:"pdate"`
	EffDate    string `csv:"effdate"`
	Form       string `csv:"form"`
	Filed      string `csv:"filed"`
	Accepted   string `csv:"accepted"`
	Instance   string `csv:"instance"`
	NCIKs      string `csv:"nciks"`
	ACIKs      string `csv:"aciks"`
}

type MFDTAG struct {
	Tag      string `csv:"tag"`
	Version  string `csv:"version"`
	Custom   string `csv:"custom"`
	Abstract string `csv:"abstract"`
	Datatype string `csv:"datatype"`
	Lord     string `csv:"lord"`
	Tlabel   string `csv:"tlabel"`
	Doc      string `csv:"doc"`
}

type MFDLAB struct {
	Adsh         string `csv:"adsh"`
	Tag          string `csv:"tag"`
	Version      string `csv:"version"`
	Std          string `csv:"std"`
	Terse        string `csv:"terse"`
	Verbose      string `csv:"verbose_val"`
	Total        string `csv:"total"`
	Negated      string `csv:"negated"`
	NegatedTerse string `csv:"negatedTerse"`
}

type MFDCAL struct {
	Adsh     string `csv:"adsh"`
	Grp      string `csv:"grp"`
	Arc      string `csv:"arc"`
	Negative string `csv:"negative"`
	PTag     string `csv:"pTag"`
	PVersion string `csv:"pversion"`
	CTag     string `csv:"ctag"`
	CVersion string `csv:"cversion"`
}

type MFDNUM struct {
	Adsh      string `csv:"adsh"`
	Tag       string `csv:"tag"`
	Version   string `csv:"version"`
	DDate     string `csv:"ddate"`
	UOM       string `csv:"uom"`
	Series    string `csv:"series"`
	Class     string `csv:"class"`
	Measure   string `csv:"measure"`
	Document  string `csv:"document"`
	OtherDims string `csv:"otherdims"`
	IPRX      string `csv:"iprx"`
	Value     string `csv:"value"`
	Footnote  string `csv:"footnote"`
	FootLen   string `csv:"footlen"`
	Dimn      string `csv:"dimn"`
	Dcml      string `csv:"dcml"`
}

type MFDTXT struct {
	Adsh      string `csv:"adsh"`
	Tag       string `csv:"tag"`
	Version   string `csv:"version"`
	DDate     string `csv:"ddate"`
	Lang      string `csv:"lang"`
	Series    string `csv:"series"`
	Class     string `csv:"class"`
	Measure   string `csv:"measure"`
	Document  string `csv:"document"`
	OtherDims string `csv:"otherdims"`
	IPRX      string `csv:"iprx"`
	Dcml      string `csv:"dcml"`
	Escaped   string `csv:"escaped"`
	SrcLen    string `csv:"srclen"`
	TxtLen    string `csv:"txtlen"`
	Footnote  string `csv:"footnote"`
	FootLen   string `csv:"footlen"`
	Context   string `csv:"context"`
	Value     string `csv:"value"`
}

type FSDSSUB struct {
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

type FSDSNUM struct {
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

type FSDSTAG struct {
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

type FSDSPRE struct {
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

func DownloadFinancialStatementDataSets(db *sqlx.DB, s *sec.SEC) error {
	if s.Verbose {
		log.Info("Downloading financial statement data sets...")
	}

	return downloadSECData(db, s, "fsds")
}

func DownloadMutualFundData(db *sqlx.DB, s *sec.SEC) error {
	if s.Verbose {
		log.Info("Downloading mutual funds data...")
	}

	return downloadSECData(db, s, "mfd")
}

func GetDataFilePath(baseURL string, yearQuarter string, fileType string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	var filePath string
	switch fileType {
	case "fsds":
		filePath = fmt.Sprintf("/files/dera/data/financial-statement-data-sets/%v.zip", yearQuarter)
	case "mfd":
		filePath = fmt.Sprintf("/files/dera/data/mutual-fund-prospectus-risk/return-summary-data-sets/%v_rr1.zip", yearQuarter)
	default:
		return "", nil
	}

	pathURL, err := url.Parse(filePath)
	if err != nil {
		return "", err
	}
	return parsedURL.ResolveReference(pathURL).String(), nil
}

func downloadSECData(db *sqlx.DB, s *sec.SEC, dataType string) error {
	worklist, err := secworklist.WillDownloadGet(db, true)
	if err != nil {
		return err
	}

	downloader := download.NewDownloader(s.Config)
	downloader.IsContentLength = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.CurrentDownloadCount = 0
	downloader.TotalDownloadsCount = len(worklist)

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}
	for _, v := range worklist {
		quarter := secutil.QuarterFromMonth(v.Month)

		if !secutil.IsCurrentYearQuarterCorrect(v.Year, quarter) {
			continue
		}

		yearQuarter := fmt.Sprintf("%vq%v", v.Year, quarter)

		fileURL, err := GetDataFilePath(s.BaseURL, yearQuarter, dataType)
		if err != nil {
			return err
		}

		if s.Verbose {
			log.Info(fmt.Sprintf("Checking file '%v' in disk: ", filepath.Base(fileURL)))
		}
		isFileCorrect, err := downloader.FileCorrect(db, fileURL, 0, "")
		if err != nil {
			return err
		}
		if s.Verbose && isFileCorrect {
			log.Info("\u2713")
		}

		if !isFileCorrect {
			if s.Verbose {
				log.Info("Downloading file...: ")
			}
			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			if s.Verbose {
				log.Info(time.Now().Format("2006-01-02 03:04:05"))
			}
			time.Sleep(rateLimit)
		}

		downloader.CurrentDownloadCount += 1
	}
	return nil
}

func IndexFinancialStatementDataSets(s *sec.SEC, db *sqlx.DB) error {
	filesPath := filepath.Join(s.Config.Main.CacheDir, "files/dera/data/financial-statement-data-sets/")
	err := indexData(s, db, filesPath, "fsds")
	if err != nil {
		return err
	}
	return nil
}

func IndexMutualFundData(s *sec.SEC, db *sqlx.DB) error {
	filesPath := filepath.Join(s.Config.Main.CacheDir, "files/dera/data/mutual-fund-prospectus-risk/return-summary-data-sets/")
	err := indexData(s, db, filesPath, "mfd")
	if err != nil {
		return err
	}
	return nil
}

func indexData(s *sec.SEC, db *sqlx.DB, filesPath string, filesType string) error {
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

		err = zipFileUpsert(s, db, filesPath, reader.File, filesType)
		if err != nil {
			secevent.CreateIndexEvent(db, filesPath, "failed", "error_inserting_secdata_in_database")
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

func zipFileUpsert(s *sec.SEC, db *sqlx.DB, pathname string, files []*zip.File, dataType string) error {
	for _, file := range files {
		fileName := strings.ToLower(file.Name)

		if fileName == "readme.htm" || fileName == "readme.html" || fileName == "rr1-metadata.json" {
			continue
		}

		upsertFunc := getDataTypeInsertFunc(fileName, dataType)
		if upsertFunc == nil {
			return fmt.Errorf("could_not_identify_file_type_func %v", fileName)
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

func getDataTypeInsertFunc(fileName string, dataType string) func(*sec.SEC, *sqlx.DB, io.ReadCloser) error {
	if dataType == "fsds" {
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

	if dataType == "mfd" {
		switch fileName {
		case "sub.txt", "sub.tsv":
			return MFDSubDataUpsert
		case "tag.txt", "tag.tsv":
			return MFDTagDataUpsert
		case "num.txt", "num.tsv":
			return MFDNumDataUpsert
		case "cal.txt", "cal.tsv":
			return MFDCalDataUpsert
		case "lab.txt", "lab.tsv":
			return MFDLabDataUpsert
		case "txt.txt", "txt.tsv":
			return MFDTxtDataUpsert
		default:
			return nil
		}
	}

	return nil
}

func MFDSubDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	subs := []MFDSUB{}
	err := gocsv.Unmarshal(reader, &subs)
	if err != nil {
		return err
	}

	for _, v := range subs {
		_, err = db.Exec(`
		INSERT INTO mfd.sub (adsh, cik, name, countryba, stprba, cityba, zipba, bas1, bas2, baph, countryma, strpma, cityma, zipma, mas1, mas2, countryinc, stprinc, ein, former, changed, fye, pdate, effdate, form, filed, accepted, instance, nciks, aciks, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, NOW(), NOW()) 
		ON CONFLICT (adsh) 
		DO NOTHING;`, v.Adsh, v.CIK, v.Name, v.CountryBA, v.StprBA, v.CityBA, v.ZIPBA, v.BAS1, v.BAS2, v.BAPH, v.CountryMA, v.StrpMA, v.CityMA, v.ZIPMA, v.MAS1, v.MAS2, v.CountryInc, v.StprInc, v.EIN, v.Former, v.Changed, v.Fye, v.PDate, v.EffDate, v.Form, v.Filed, v.Accepted, v.Instance, v.NCIKs, v.ACIKs)
		if err != nil {
			return err
		}
	}
	return nil
}

func MFDTagDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	tags := []MFDTAG{}
	err := gocsv.Unmarshal(reader, &tags)
	if err != nil {
		return err
	}

	for _, v := range tags {
		_, err = db.Exec(`
			INSERT INTO mfd.tag (tag, version, custom, abstract, datatype, lord, tlabel, doc, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) 
			ON CONFLICT (tag, version) 
			DO NOTHING;`, v.Tag, v.Version, v.Custom, v.Abstract, v.Datatype, v.Lord, v.Tlabel, v.Doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func MFDLabDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})
	labs := []MFDLAB{}

	err := gocsv.Unmarshal(reader, &labs)
	if err != nil {
		return err
	}

	for _, v := range labs {
		_, err = db.Exec(`
			INSERT INTO mfd.lab (adsh, tag, version, std, terse, verbose_val, total, negated, negatedterse, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, tag, version)
			DO NOTHING;`, v.Adsh, v.Tag, v.Version, v.Std, v.Terse, v.Verbose, v.Total, v.Negated, v.NegatedTerse)
		if err != nil {
			return err
		}
	}
	return nil
}

func MFDCalDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	cals := []MFDCAL{}
	err := gocsv.Unmarshal(reader, &cals)
	if err != nil {
		return err
	}

	for _, v := range cals {
		_, err = db.Exec(`
			INSERT INTO mfd.cal (adsh, grp, arc, negative, ptag, pversion, ctag, cversion, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) 
			ON CONFLICT (adsh, grp, arc) 
			DO NOTHING;`, v.Adsh, v.Grp, v.Arc, v.Negative, v.PTag, v.PVersion, v.CTag, v.CVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

func MFDNumDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	nums := []MFDNUM{}
	err := gocsv.Unmarshal(reader, &nums)
	if err != nil {
		return err
	}

	for _, v := range nums {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM mfd.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO mfd.num (adsh, tag, version, ddate, uom, series, class, measure, document, otherdims, iprx, value, footnote, footlen, dimn, dcml, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()) 
			ON CONFLICT (adsh, tag, version, ddate, uom, series, class, measure, document, otherdims, iprx) 
			DO NOTHING;`, v.Adsh, v.Tag, v.Version, v.DDate, v.UOM, v.Series, v.Class, v.Measure, v.Document, v.OtherDims, v.IPRX, v.Value, v.Footnote, v.FootLen, v.Dimn, v.Dcml)
		if err != nil {
			return err
		}
	}
	return nil
}

func MFDTxtDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	txt := []MFDTXT{}
	err := gocsv.Unmarshal(reader, &txt)
	if err != nil {
		return err
	}

	for _, v := range txt {
		_, err = db.Exec(`
			INSERT INTO mfd.txt (adsh, tag, version, ddate, lang, series, class, measure, document, otherdims, iprx, dcml, escaped, srclen, txtlen, footnote, footlen, context, value, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW(), NOW()) 
			ON CONFLICT (adsh, tag, version, ddate, series, class, measure, document, otherdims, iprx) 
			DO NOTHING;`, v.Adsh, v.Tag, v.Version, v.DDate, v.Lang, v.Series, v.Class, v.Measure, v.Document, v.OtherDims, v.IPRX, v.Dcml, v.Escaped, v.SrcLen, v.TxtLen, v.Footnote, v.FootLen, v.Context, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func FSDSSubDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

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
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

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
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

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
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

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
