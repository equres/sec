package secdata

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
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

type SECDataOps interface {
	GetDataType() string
	GetDataFilePath(baseURL string, yearQuarter string) (string, error)
	GetDataDirPath() string
	GetDataTypeInsertFunc(fileName string) func(*sec.SEC, *sqlx.DB, io.ReadCloser) error
}

type SECData struct {
	SECDataOps SECDataOps
}

func NewSECData(s SECDataOps) *SECData {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})
	return &SECData{
		SECDataOps: s,
	}
}

func (sd *SECData) DownloadSECData(db *sqlx.DB, s *sec.SEC) error {
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

		fileURL, err := sd.SECDataOps.GetDataFilePath(s.BaseURL, yearQuarter)
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

func (sd *SECData) IndexData(s *sec.SEC, db *sqlx.DB) error {
	filesPath := filepath.Join(s.Config.Main.CacheDir, sd.SECDataOps.GetDataDirPath())
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

		err = sd.ZIPFileUpsert(s, db, filesPath, reader.File)
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

func (sd *SECData) ZIPFileUpsert(s *sec.SEC, db *sqlx.DB, pathname string, files []*zip.File) error {
	for _, file := range files {
		fileName := strings.ToLower(file.Name)

		if fileName == "readme.htm" || fileName == "readme.html" || fileName == "rr1-metadata.json" {
			continue
		}

		upsertFunc := sd.SECDataOps.GetDataTypeInsertFunc(fileName)
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
