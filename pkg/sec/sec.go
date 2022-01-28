// Copyright (c) 2021 Koszek Systems. All rights reserved.

package sec

import (
	"database/sql"
	"encoding/xml"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/cache"
	"github.com/equres/sec/pkg/config"
	"github.com/jmoiron/sqlx"
)

type RSSFile struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}
type Channel struct {
	Text          string `xml:",chardata"`
	Title         string `xml:"title"`
	Link          Link   `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	PubDate       string `xml:"pubDate"`
	LastBuildDate string `xml:"lastBuildDate"`
	Item          []Item `xml:"item"`
}

type Link struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Atom string `xml:"atom,attr"`
}

type Item struct {
	Text        string     `xml:",chardata"`
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Guid        string     `xml:"guid"`
	Enclosure   Enclosure  `xml:"enclosure"`
	Description string     `xml:"description"`
	PubDate     string     `xml:"pubDate"`
	XbrlFiling  XbrlFiling `xml:"xbrlFiling"`
}

type Enclosure struct {
	Text   string `xml:",chardata"`
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type XbrlFiling struct {
	Text               string    `xml:",chardata"`
	Edgar              string    `xml:"edgar,attr"`
	CompanyName        string    `xml:"companyName"`
	FormType           string    `xml:"formType"`
	FilingDate         string    `xml:"filingDate"`
	CikNumber          string    `xml:"cikNumber"`
	AccessionNumber    string    `xml:"accessionNumber"`
	FileNumber         string    `xml:"fileNumber"`
	AcceptanceDatetime string    `xml:"acceptanceDatetime"`
	Period             string    `xml:"period"`
	AssistantDirector  string    `xml:"assistantDirector"`
	AssignedSic        string    `xml:"assignedSic"`
	FiscalYearEnd      string    `xml:"fiscalYearEnd"`
	XbrlFiles          XbrlFiles `xml:"xbrlFiles"`
}

type XbrlFiles struct {
	Text     string     `xml:",chardata"`
	XbrlFile []XbrlFile `xml:"xbrlFile"`
}

type XbrlFile struct {
	Text        string `xml:",chardata"`
	Sequence    string `xml:"sequence,attr"`
	File        string `xml:"file,attr"`
	Type        string `xml:"type,attr"`
	Size        string `xml:"size,attr"`
	Description string `xml:"description,attr"`
	InlineXBRL  string `xml:"inlineXBRL,attr"`
	URL         string `xml:"url,attr"`
}

type Entry struct {
	URL  string
	Path string
	Size int
}

type SECItemFile struct {
	ID                 int            `db:"id"`
	Ticker             sql.NullString `db:"ticker"`
	Title              string         `db:"title"`
	Link               string         `db:"link"`
	Guid               string         `db:"guid"`
	EnclosureURL       string         `db:"enclosure_url"`
	EnclosureLength    int            `db:"enclosure_length"`
	EnclosureType      string         `db:"enclosure_type"`
	Description        string         `db:"description"`
	PubDate            time.Time      `db:"pubdate"`
	CompanyName        string         `db:"companyname"`
	FormType           string         `db:"formtype"`
	FillingDate        time.Time      `db:"fillingdate"`
	CIKNumber          string         `db:"ciknumber"`
	AccessionNumber    string         `db:"accessionnumber"`
	FileNumber         string         `db:"filenumber"`
	AcceptanceDatetime string         `db:"acceptancedatetime"`
	Period             string         `db:"period"`
	AssistantDirector  string         `db:"assistantdirector"`
	AssignedSic        int            `db:"assignedsic"`
	FiscalYearEnd      int            `db:"fiscalyearend"`
	XbrlSequence       string         `db:"xbrlsequence"`
	XbrlFile           string         `db:"xbrlfile"`
	XbrlType           string         `db:"xbrltype"`
	XbrlSize           int            `db:"xbrlsize"`
	XbrlDescription    string         `db:"xbrldescription"`
	XbrlInlineXbrl     bool           `db:"xbrlinlinexbrl"`
	XbrlURL            string         `db:"xbrlurl"`
	XbrlBody           string         `db:"xbrlbody"`
}

type Company struct {
	CompanyName string
	CIKNumber   string
}

type SIC struct {
	SIC    string `db:"SIC"`
	Office string `db:"office"`
	Title  string `db:"title"`
}

// Ticker Struct Based on JSON
type SecTicker struct {
	Cik      int    `json:"cik_str"`
	Ticker   string `json:"ticker"`
	Title    string `json:"title"`
	Exchange string `json:"exchange"`
}

type SEC struct {
	BaseURL string
	Verbose bool
	Config  config.Config
	Debug   bool
	Cache   cache.Cache
}

func NewSEC(config config.Config) (*SEC, error) {
	redisCache := cache.NewCache(&config)
	return &SEC{
		BaseURL: config.Main.BaseURL,
		Config:  config,
		Cache:   redisCache,
	}, nil
}

func GetAllCompanies(db *sqlx.DB) ([]Company, error) {
	var companies []Company

	err := db.Select(&companies, "SELECT DISTINCT companyname, ciknumber FROM sec.secitemfile;")
	if err != nil {
		return nil, err
	}

	return companies, err
}

func GetCompanyFilingsFromCIK(db *sqlx.DB, cik int) (map[string][]SECItemFile, error) {
	var secItemFiles []SECItemFile

	err := db.Select(&secItemFiles, `
	SELECT * FROM (
		SELECT DISTINCT ON (accessionnumber) 
			companyname, ciknumber, formtype, fillingdate 
		FROM sec.secItemFile WHERE ciknumber = $1
		) filings 
	ORDER BY fillingdate desc;
	`, cik)
	if err != nil {
		return nil, err
	}

	filings := make(map[string][]SECItemFile)

	for _, item := range secItemFiles {
		year := strconv.Itoa(item.FillingDate.Year())
		filings[year] = append(filings[year], item)
	}

	return filings, nil
}
