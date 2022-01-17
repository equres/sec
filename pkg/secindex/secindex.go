package secindex

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secevent"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"jaytaylor.com/html2text"
)

func InsertAllSecItemFile(db *sqlx.DB, s *sec.SEC, rssFiles []sec.RSSFile, worklistMap map[string]sec.Entry, totalCount int) error {
	currentCount := 0
	for _, rssFile := range rssFiles {
		for _, v1 := range rssFile.Channel.Item {
			err := SecItemFileUpsert(db, s, v1, worklistMap, &currentCount, totalCount)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SecItemFileUpsert(db *sqlx.DB, s *sec.SEC, item sec.Item, worklist map[string]sec.Entry, currentCount *int, totalCount int) error {
	var err error

	var enclosureLength int
	if item.Enclosure.Length != "" {
		enclosureLength, err = strconv.Atoi(item.Enclosure.Length)
		if err != nil {
			return err
		}
	}
	var assignedSic int
	if item.XbrlFiling.AssignedSic != "" {
		assignedSic, err = strconv.Atoi(item.XbrlFiling.AssignedSic)
		if err != nil {
			return err
		}
	}

	var fiscalYearEnd int
	if item.XbrlFiling.FiscalYearEnd != "" {
		fiscalYearEnd, err = strconv.Atoi(item.XbrlFiling.FiscalYearEnd)
		if err != nil {
			return err
		}
	}
	var cikNumber int
	if item.XbrlFiling.CikNumber != "" {
		cikNumber, err = strconv.Atoi(item.XbrlFiling.CikNumber)
		if err != nil {
			return err
		}
	}

	for _, v := range item.XbrlFiling.XbrlFiles.XbrlFile {
		if _, ok := worklist[v.URL]; ok {
			continue
		}
		if s.Verbose {
			currentCountFloat := float64(*currentCount)
			totalCountFloat := float64(totalCount)
			percentage := (currentCountFloat / totalCountFloat) * 100

			log.Info(fmt.Sprintf("[%d/%d/%f%%] %s inserting file %v", *currentCount, totalCount, percentage, time.Now().Format("2006-01-02 03:04:05"), v.URL))
		}

		var xbrlInline bool
		if v.InlineXBRL != "" {
			xbrlInline, err = strconv.ParseBool(v.InlineXBRL)
			if err != nil {
				return err
			}
		}

		var xbrlSequence int
		if v.Sequence != "" {
			xbrlSequence, err = strconv.Atoi(v.Sequence)
			if err != nil {
				return err
			}
		}

		var xbrlSize int
		if v.Size != "" {
			xbrlSize, err = strconv.Atoi(v.Size)
			if err != nil {
				return err
			}
		}

		fileUrl, err := url.Parse(v.URL)
		if err != nil {
			return err
		}

		var fileBody string
		filePath := filepath.Join(s.Config.Main.CacheDir, fileUrl.Path)
		_, err = os.Stat(filePath)
		if err != nil {
			filePath = ""
		}

		// If filePath is empty then check unzipped files
		if filePath == "" {
			filePath = filepath.Join(s.Config.Main.CacheDirUnpacked, fileUrl.Path)
			_, err = os.Stat(filePath)
			if err != nil {
				filePath = ""
			}
		}

		if filePath != "" {
			fileBody, err = GetXbrlFileBody(filePath)
			if err != nil {
				return err
			}
		}

		if fileBody == "" && IsFileIndexable(filePath) {
			secevent.CreateIndexEvent(db, v.URL, "failed", "could_not_find_file")
		}

		_, err = db.Exec(`
		INSERT INTO sec.secItemFile (title, link, guid, enclosure_url, enclosure_length, enclosure_type, description, pubdate, companyname, formtype, fillingdate, ciknumber, accessionnumber, filenumber, acceptancedatetime, period, assistantdirector, assignedsic, fiscalyearend, xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl, xbrlbody, XbrlFilePath, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, NOW(), NOW()) 

		ON CONFLICT (xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl)
		DO UPDATE SET title=EXCLUDED.title, link=EXCLUDED.link, guid=EXCLUDED.guid, enclosure_url=EXCLUDED.enclosure_url, enclosure_length=EXCLUDED.enclosure_length, enclosure_type=EXCLUDED.enclosure_type, description=EXCLUDED.description, pubdate=EXCLUDED.pubdate, companyname=EXCLUDED.companyname, formtype=EXCLUDED.formtype, fillingdate=EXCLUDED.fillingdate, ciknumber=EXCLUDED.ciknumber, accessionnumber=EXCLUDED.accessionnumber, filenumber=EXCLUDED.filenumber, acceptancedatetime=EXCLUDED.acceptancedatetime, period=EXCLUDED.period, assistantdirector=EXCLUDED.assistantdirector, assignedsic=EXCLUDED.assignedsic, fiscalyearend=EXCLUDED.fiscalyearend, xbrlsequence=EXCLUDED.xbrlsequence, xbrlfile=EXCLUDED.xbrlfile, xbrltype=EXCLUDED.xbrltype, xbrlsize=EXCLUDED.xbrlsize, xbrldescription=EXCLUDED.xbrldescription, xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl, xbrlurl=EXCLUDED.xbrlurl, xbrlfilepath=EXCLUDED.xbrlfilepath, updated_at=NOW()
		WHERE secItemFile.xbrlsequence=EXCLUDED.xbrlsequence AND secItemFile.xbrlfile=EXCLUDED.xbrlfile AND secItemFile.xbrltype=EXCLUDED.xbrltype AND secItemFile.xbrlsize=EXCLUDED.xbrlsize AND secItemFile.xbrldescription=EXCLUDED.xbrldescription AND secItemFile.xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl AND secItemFile.xbrlurl=EXCLUDED.xbrlurl AND secItemFile.xbrlbody=EXCLUDED.xbrlbody;`,
			item.Title, item.Link, item.Guid, item.Enclosure.URL, enclosureLength, item.Enclosure.Type, item.Description, item.PubDate, item.XbrlFiling.CompanyName, item.XbrlFiling.FormType, item.XbrlFiling.FilingDate, cikNumber, item.XbrlFiling.AccessionNumber, item.XbrlFiling.FileNumber, item.XbrlFiling.AcceptanceDatetime, item.XbrlFiling.Period, item.XbrlFiling.AssistantDirector, assignedSic, fiscalYearEnd, xbrlSequence, v.File, v.Type, xbrlSize, v.Description, xbrlInline, v.URL, fileBody, filePath)
		if err != nil {
			secevent.CreateIndexEvent(db, filePath, "failed", "error_inserting_in_database")
			return err
		}

		*currentCount++

		secevent.CreateIndexEvent(db, filePath, "success", "")
	}
	return nil
}

func IsFileIndexable(filename string) bool {
	fileExtension := strings.ToLower(filepath.Ext(filename))

	if fileExtension == ".html" || fileExtension == ".htm" || fileExtension == ".xml" {
		return true
	}
	return false
}

func IsFileTypeHTML(filename string) bool {
	fileExtension := strings.ToLower(filepath.Ext(filename))

	if fileExtension == ".html" || fileExtension == ".htm" {
		return true
	}
	return false
}

func GetXbrlFileBody(filePath string) (string, error) {
	xbrlFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer xbrlFile.Close()

	data, err := ioutil.ReadAll(xbrlFile)
	if err != nil {
		return "", err
	}

	var fileBody string

	if IsFileIndexable(filePath) {
		fileBody = string(data)
		if IsFileTypeHTML(filePath) {
			fileBody, err = html2text.FromString(string(data))
			if err != nil {
				return "", err
			}
		}
	}
	return fileBody, nil
}

func GetXbrlFileBodyFromZIPFile(currentFile *zip.File, filePath string) (string, error) {
	if currentFile == nil {
		return "", nil
	}

	fileReader, err := currentFile.Open()
	if err != nil {
		return "", err
	}

	stringBuilder := new(strings.Builder)
	_, err = io.Copy(stringBuilder, fileReader)
	if err != nil {
		return "", err
	}

	var fileBody string
	if IsFileIndexable(currentFile.Name) {
		fileBody = stringBuilder.String()
		if IsFileTypeHTML(filePath) {
			fileBody, err = html2text.FromString(stringBuilder.String())
			if err != nil {
				return "", err
			}
		}
	}

	return fileBody, nil
}

func ZIPContentUpsert(db *sqlx.DB, pathname string, files []*zip.File) error {
	// Keeping only directories
	dirsPath := filepath.Dir(pathname)

	// Spliting directories
	dirs := strings.Split(dirsPath, "\\")
	if len(dirs) == 1 {
		dirs = strings.Split(dirsPath, "/")
	}

	// Keeping only CIK and Accession Number
	dirs = dirs[len(dirs)-2:]

	cik := dirs[0]
	accession := dirs[1]

	for _, file := range files {
		reader, err := file.Open()
		if err != nil {
			return err
		}

		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(reader)
		if err != nil {
			return err
		}

		reader.Close()

		var xbrlBody string

		if IsFileIndexable(file.FileInfo().Name()) {
			xbrlBody = buf.String()
		}

		_, err = db.Exec(`
			INSERT INTO sec.secItemFile (ciknumber, accessionnumber, xbrlfile, xbrlsize, xbrlbody, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) 
			ON CONFLICT (cikNumber, accessionNumber, xbrlFile, xbrlSize)
			DO NOTHING;`, cik, accession, file.Name, int(file.FileInfo().Size()), xbrlBody)
		if err != nil {
			secevent.CreateIndexEvent(db, pathname, "failed", "indexz_error_inserting_in_database")
		}
	}

	secevent.CreateIndexEvent(db, pathname, "success", "")

	return nil
}

func IndexZIPFileContent(db *sqlx.DB, s *sec.SEC, rssFile sec.RSSFile, worklist []secworklist.Worklist) error {
	totalCount := len(rssFile.Channel.Item)
	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		if v1.Enclosure.URL == "" {
			continue
		}
		parsedURL, err := url.Parse(v1.Enclosure.URL)
		if err != nil {
			return err
		}
		zipPath := parsedURL.Path

		zipCachePath := filepath.Join(s.Config.Main.CacheDir, zipPath)
		_, err = os.Stat(zipCachePath)
		if err != nil {
			secevent.CreateIndexEvent(db, zipCachePath, "failed", "zip_file_does_not_exist")
			log.Info("please run sec dowz to download all ZIP files then run sec indexz again to index them")
			return err
		}

		reader, err := zip.OpenReader(zipCachePath)
		if err != nil {
			secevent.CreateIndexEvent(db, zipCachePath, "failed", "corrupt_zip_file")
			log.Errorf("Could not access the file %v", zipCachePath)
			continue
		}

		err = ZIPContentUpsert(db, zipPath, reader.File)
		if err != nil {
			secevent.CreateIndexEvent(db, zipCachePath, "failed", "indexz_error_inserting_in_database")
			return err
		}

		secevent.CreateIndexEvent(db, zipCachePath, "success", "")

		reader.Close()
		currentCount++

		if s.Verbose {
			log.Info(fmt.Sprintf("[%d/%d] %s inserted for current file...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05")))
		}
	}
	return nil
}

func IndexSICCodes(s *sec.SEC, db *sqlx.DB) error {
	sicList, err := secutil.GetSICCodes(s, db)
	if err != nil {
		return err
	}

	for _, sic := range sicList {
		_, err = db.Exec(`
		INSERT INTO sec.sics (sic, office, title, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW()) 

		ON CONFLICT (sic)
		DO UPDATE SET sic=EXCLUDED.sic, office=EXCLUDED.office, title=EXCLUDED.title, updated_at=NOW()
		WHERE sics.sic=EXCLUDED.sic AND sics.office=EXCLUDED.office AND sics.title=EXCLUDED.title;`,
			sic.SIC, sic.Office, sic.Title)
		if err != nil {
			eventErr := database.CreateOtherEvent(db, "index", "sic", "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}
	eventErr := database.CreateOtherEvent(db, "index", "sic", "success")
	if eventErr != nil {
		return eventErr
	}

	return nil
}
