package secdata

import (
	"fmt"
	"io"
	"net/url"

	"github.com/equres/sec/pkg/sec"
	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
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

type SECDataOpsMFD struct {
	DataType string
}

func NewSECDataOpsMFD() *SECDataOpsMFD {
	return &SECDataOpsMFD{
		DataType: "mfd",
	}
}

func (s *SECDataOpsMFD) GetDataType() string {
	return s.DataType
}

func (s *SECDataOpsMFD) GetDataFilePath(baseURL string, yearQuarter string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("/files/dera/data/mutual-fund-prospectus-risk/return-summary-data-sets/%v_rr1.zip", yearQuarter)

	pathURL, err := url.Parse(filePath)
	if err != nil {
		return "", err
	}
	return parsedURL.ResolveReference(pathURL).String(), nil
}

func (s *SECDataOpsMFD) GetDataDirPath() string {
	return "files/dera/data/mutual-fund-prospectus-risk/return-summary-data-sets/"
}

func (s *SECDataOpsMFD) GetDataTypeInsertFunc(fileName string) func(*sec.SEC, *sqlx.DB, io.ReadCloser) error {
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

func MFDSubDataUpsert(s *sec.SEC, db *sqlx.DB, reader io.ReadCloser) error {
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
