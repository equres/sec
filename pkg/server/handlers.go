package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	humanize "github.com/dustin/go-humanize"
	"github.com/equres/sec/pkg/cache"
	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/seccik"
	"github.com/equres/sec/pkg/secsic"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
)

func (s Server) GenerateRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	assets, err := fs.Sub(GlobalAssetsFS, "_assets")
	if err != nil {
		return nil, err
	}
	router.PathPrefix("/_assets").Handler(http.StripPrefix("/_assets", http.FileServer(http.FS(assets))))

	router.HandleFunc("/", s.HandlerHome).Methods("GET")
	router.HandleFunc("/about", s.HandlerAbout).Methods("GET")
	router.HandleFunc("/help", s.HandlerHelp).Methods("GET")
	router.HandleFunc("/signup", s.HandlerSignUp).Methods("GET")
	router.HandleFunc("/filings/{year}", s.HandlerMonthsPage).Methods("GET")
	router.HandleFunc("/filings/{year}/{month}", s.HandlerDaysPage).Methods("GET")
	router.HandleFunc("/filings/{year}/{month}/{day}", s.HandlerCompaniesPage).Methods("GET")
	router.HandleFunc("/filings/{year}/{month}/{day}/{cik}", s.HandlerFilingsPage).Methods("GET")
	router.HandleFunc("/company", s.HandlerCompaniesListPage).Methods("GET")
	router.HandleFunc("/company/{companySlug}", s.HandlerCompanyFilingsPage).Methods("GET")
	router.HandleFunc("/sic", s.HandlerSICListPage).Methods("GET")
	router.HandleFunc("/sic/{sic}", s.HandlerSICCompaniesPage).Methods("GET")
	router.HandleFunc("/stats", s.HandlerStatsPage).Methods("GET")
	router.HandleFunc("/api/v1/uptime", s.HandlerUptime).Methods("GET")
	router.HandleFunc("/robots.txt", s.HanderRobots).Methods("GET")
	router.PathPrefix("/").HandlerFunc(s.HandlerFiles)
	return router, nil
}

func (s Server) HandlerHome(w http.ResponseWriter, r *http.Request) {
	// Get top five filings from redis cache
	recentFilingsFormattedJSON, err := s.Cache.Get(cache.SECTopFiveRecentFilings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type FormattedFiling struct {
		CompanyName string
		FillingDate string
		FormType    string
		XbrlURL     string
	}

	var recentFilingsFormatted []FormattedFiling

	err = json.Unmarshal([]byte(recentFilingsFormattedJSON), &recentFilingsFormatted)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ciksCount, err := s.Cache.Get(cache.SECCIKsCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filesCount, err := s.Cache.Get(cache.SECFilesCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companiesCount, err := s.Cache.Get(cache.SECCompaniesCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})

	content["UniqueCIKsCount"] = ciksCount
	content["UniqueFilesCount"] = filesCount
	content["UniqueCompaniesCount"] = companiesCount
	content["RecentFilings"] = recentFilingsFormatted
	content["Years"], err = secworklist.UniqueYears(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.RenderTemplate(w, "index.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerAbout(w http.ResponseWriter, r *http.Request) {
	content := make(map[string]interface{})
	err := s.RenderTemplate(w, "about.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerHelp(w http.ResponseWriter, r *http.Request) {
	content := make(map[string]interface{})
	err := s.RenderTemplate(w, "help.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	content := make(map[string]interface{})
	err := s.RenderTemplate(w, "signup.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerFiles(w http.ResponseWriter, r *http.Request) {
	filename := strings.ReplaceAll(r.URL.Path, "/static/", "")

	filePath := filepath.Join(s.Config.Main.CacheDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = filepath.Join(s.Config.Main.CacheDirUnpacked, filename)
	}

	http.ServeFile(w, r, filePath)
}

func (s Server) HandlerMonthsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	year, err := getIntVar(vars, "year")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	monthsJSON, err := s.Cache.MustGet(fmt.Sprintf("%v_%v", cache.SECMonthsInYear, year))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var months []int
	err = json.Unmarshal([]byte(monthsJSON), &months)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Year"] = year
	content["Months"] = months

	err = s.RenderTemplate(w, "months.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) HandlerDaysPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	year, err := getIntVar(vars, "year")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	month, err := getIntVar(vars, "month")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	daysJSON, err := s.Cache.MustGet(fmt.Sprintf("%v_%v_%v", cache.SECDaysInMonth, year, month))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var days []int
	err = json.Unmarshal([]byte(daysJSON), &days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allDays := make(map[int]bool)
	for _, day := range days {
		allDays[day] = true
	}

	daysCountForMonth := time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC).Day()
	for i := 1; i <= daysCountForMonth; i++ {
		if _, ok := allDays[i]; !ok {
			allDays[i] = false
		}
	}

	var monthString time.Month = time.Month(month)

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["MonthString"] = monthString.String()
	content["Days"] = allDays

	err = s.RenderTemplate(w, "days.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) HandlerCompaniesPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	year, err := getIntVar(vars, "year")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	month, err := getIntVar(vars, "month")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	day, err := getIntVar(vars, "day")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companiesJSON, err := s.Cache.MustGet(fmt.Sprintf("%v_%v_%v_%v", cache.SECCompaniesInDay, year, month, day))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var companies []sec.SECItemFile
	err = json.Unmarshal([]byte(companiesJSON), &companies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var monthString time.Month = time.Month(month)
	var dayString string = time.Weekday(day % 7).String()
	var dayOrdinal string = humanize.Ordinal(day)

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["MonthString"] = monthString.String()
	content["Day"] = day
	content["DayString"] = dayString
	content["DayOrdinal"] = dayOrdinal
	content["Companies"] = companies

	err = s.RenderTemplate(w, "companies.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) HandlerFilingsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	year, err := getIntVar(vars, "year")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	month, err := getIntVar(vars, "month")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	day, err := getIntVar(vars, "day")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cik, err := getIntVar(vars, "cik")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filingsJSON, err := s.Cache.MustGet(fmt.Sprintf("%v_%v_%v_%v_%v", cache.SECFilingsInDay, year, month, day, cik))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var filings []sec.SECItemFile
	err = json.Unmarshal([]byte(filingsJSON), &filings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companyName, err := seccik.GetCompanyNameFromCIK(s.DB, cik)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var monthString time.Month = time.Month(month)
	var dayString string = time.Weekday(day % 7).String()
	var dayOrdinal string = humanize.Ordinal(day)

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["MonthString"] = monthString.String()
	content["Day"] = day
	content["DayString"] = dayString
	content["DayOrdinal"] = dayOrdinal
	content["CompanyName"] = companyName
	content["CIK"] = cik
	content["Filings"] = filings

	err = s.RenderTemplate(w, "filings.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) HandlerCompaniesListPage(w http.ResponseWriter, r *http.Request) {

	companiesJSON, err := s.Cache.MustGet(cache.SECCompanySlugs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companies := make(map[string]sec.Company)
	err = json.Unmarshal([]byte(companiesJSON), &companies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Companies"] = companies

	err = s.RenderTemplate(w, "companieslist.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerCompanyFilingsPage(w http.ResponseWriter, r *http.Request) {
	companies, err := sec.GetAllCompanies(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	companySlug := vars["companySlug"]
	company := GetCompanyFromSlug(companies, companySlug)

	cik, err := strconv.Atoi(company.CIKNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type FormattedFiling struct {
		CompanyName string
		FillingDate string
		FormType    string
		FilingsURL  string
	}

	type YearFilings struct {
		Year    string
		Filings []FormattedFiling
	}

	filingsJSON, err := s.Cache.MustGet(fmt.Sprintf("%v_%v", cache.SECCompanyFilings, cik))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var filings []YearFilings
	err = json.Unmarshal([]byte(filingsJSON), &filings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companyName, err := seccik.GetCompanyNameFromCIK(s.DB, cik)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Filings"] = filings
	content["CompanyName"] = companyName

	err = s.RenderTemplate(w, "companyfilings.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerSICListPage(w http.ResponseWriter, r *http.Request) {
	sicJSON, err := s.Cache.MustGet(cache.SECSICs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sics []sec.SIC
	err = json.Unmarshal([]byte(sicJSON), &sics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["SICs"] = sics

	err = s.RenderTemplate(w, "sicslist.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerSICCompaniesPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sic := vars["sic"]

	companiesJSON, err := s.Cache.MustGet(fmt.Sprintf("%v_%v", cache.SECCompaniesWithSIC, sic))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var companies []sec.Company
	err = json.Unmarshal([]byte(companiesJSON), &companies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	categoryName, err := secsic.GetCategoryNameFromSIC(s.DB, sic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Companies"] = GetCompanySlugs(companies)
	content["CategoryName"] = categoryName

	err = s.RenderTemplate(w, "companieslist.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerStatsPage(w http.ResponseWriter, r *http.Request) {
	allStats, err := s.GetStatsFromRedis(s.Config.Redis)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	content := make(map[string]interface{})
	content["EventStats"] = allStats

	err = s.RenderTemplate(w, "stats.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) HandlerUptime(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK: ", GlobalUptime)
}

func (s Server) HanderRobots(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `User-agent: MJ12bot
Disallow: /

User-agent: *
Disallow: /signup

User-agent: *
Allow: /

Sitemap: https://equres.com/_cache/sitemap.xml
	`)
}

func (s Server) RenderTemplate(w http.ResponseWriter, tmplName string, data interface{}) error {
	funcMap := template.FuncMap{
		"formatAccession": func(accession string) string {
			return strings.ReplaceAll(accession, "-", "")
		},
		"AppVersion": func() string {
			return fmt.Sprintf("%s %s", s.SHA1Ver, s.BuildTime)
		},
		"Increment": func(i int) int {
			return i + 1
		},
	}

	tmpl, err := template.New("tmpl").Funcs(funcMap).ParseFS(s.TemplatesFS, "templates/"+tmplName, "templates/base.layout.gohtml")
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		return err
	}

	return nil
}

func getIntVar(vars map[string]string, varName string) (int, error) {
	varStr, ok := vars[varName]
	if !ok {
		log.Info("please choose a proper year and month")
		return 0, errors.New("please choose a proper year and month")
	}
	value, err := strconv.Atoi(varStr)
	if err != nil {
		log.Info("please choose a proper year and month")
		return 0, errors.New("please choose a proper year and month")
	}
	return value, nil
}

func GetCompanySlugs(companies []sec.Company) map[string]sec.Company {
	companySlugs := make(map[string]sec.Company)
	for _, company := range companies {
		companySlugs[slug.Make(company.CompanyName)] = company
	}
	return companySlugs
}

func GetCompanyFromSlug(companies []sec.Company, companySlug string) sec.Company {
	companySlugs := GetCompanySlugs(companies)
	if company, ok := companySlugs[companySlug]; ok {
		return company
	}
	return sec.Company{}
}

func (s Server) GetStatsFromRedis(redisConfig config.RedisConfig) (map[string]int, error) {
	allStatsJSON, err := s.Cache.Get(cache.SECCacheStats)
	if err != nil {
		return nil, err
	}

	allStats := make(map[string]int)

	err = json.Unmarshal([]byte(allStatsJSON), &allStats)
	if err != nil {
		return nil, err
	}

	return allStats, nil
}
