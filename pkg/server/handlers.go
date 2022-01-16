package server

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	humanize "github.com/dustin/go-humanize"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/seccik"
	"github.com/equres/sec/pkg/secevent"
	"github.com/equres/sec/pkg/secutil"
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
	router.HandleFunc("/filings/{year}", s.HandlerMonthsPage).Methods("GET")
	router.HandleFunc("/filings/{year}/{month}", s.HandlerDaysPage).Methods("GET")
	router.HandleFunc("/filings/{year}/{month}/{day}", s.HandlerCompaniesPage).Methods("GET")
	router.HandleFunc("/filings/{year}/{month}/{day}/{cik}", s.HandlerFilingsPage).Methods("GET")
	router.HandleFunc("/company", s.HandlerCompaniesListPage).Methods("GET")
	router.HandleFunc("/company/{companySlug}", s.HandlerCompanyFilingsPage).Methods("GET")
	router.HandleFunc("/stats", s.HandlerStatsPage).Methods("GET")
	router.HandleFunc("/api/v1/uptime", s.HandlerUptime).Methods("GET")
	router.PathPrefix("/").HandlerFunc(s.HandlerFiles)
	return router, nil
}

func (s Server) HandlerHome(w http.ResponseWriter, r *http.Request) {
	content := make(map[string]interface{})
	recentFilings, err := secutil.GetFiveRecentFilings(s.DB)
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

	for _, filing := range recentFilings {
		filingURL := fmt.Sprintf("/filings/%v/%v/%v/%v", filing.FillingDate.Year(), int(filing.FillingDate.Month()), filing.FillingDate.Day(), filing.CIKNumber)

		formattedFiling := FormattedFiling{
			CompanyName: filing.CompanyName,
			FillingDate: filing.FillingDate.Format("2006-01-02"),
			FormType:    filing.FormType,
			XbrlURL:     filingURL,
		}
		recentFilingsFormatted = append(recentFilingsFormatted, formattedFiling)
	}
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

	content := make(map[string]interface{})
	content["Year"] = year
	content["Months"], err = secworklist.MonthsInYear(s.DB, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	days, err := secutil.GetFilingDaysFromMonthYear(s.DB, year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var monthString time.Month = time.Month(month)

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["MonthString"] = monthString.String()
	content["Days"] = days

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

	companies, err := secutil.GetFilingCompaniesFromYearMonthDay(s.DB, year, month, day)
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

	filings, err := secutil.SearchFilingsByYearMonthDayCIK(s.DB, year, month, day, cik)
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
	companies, err := sec.GetAllCompanies(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Companies"] = GetCompanySlugs(companies)

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

	filings, err := sec.GetCompanyFilingsFromCIK(s.DB, cik)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companyName, err := seccik.GetCompanyNameFromCIK(s.DB, cik)
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

	formattedFilings := make(map[string][]FormattedFiling)

	for year, secItemFiles := range filings {
		for _, item := range secItemFiles {
			fileLink, err := url.Parse(item.XbrlURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			formType := secutil.GetFullFormType(item.FormType)
			formattedFilings[year] = append(formattedFilings[year], FormattedFiling{
				CompanyName: item.CompanyName,
				FillingDate: item.FillingDate.Format("2006-01-02"),
				FormType:    formType,
				XbrlURL:     fileLink.Path,
			})
		}
	}

	content := make(map[string]interface{})
	content["Filings"] = formattedFilings
	content["CompanyName"] = companyName

	err = s.RenderTemplate(w, "companyfilings.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s Server) HandlerStatsPage(w http.ResponseWriter, r *http.Request) {
	eventStatsArr, err := secevent.GetEventStats(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	allStats := make(map[string]int)
	for _, event := range eventStatsArr {
		statValue := 2
		if event.FilesBroken > 0 {
			statValue--
		}

		if event.FilesIndexed == 0 || event.FilesDownloaded == 0 {
			statValue--
		}

		allStats[event.Date] = statValue
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

func (s Server) RenderTemplate(w http.ResponseWriter, tmplName string, data interface{}) error {
	funcMap := template.FuncMap{
		"formatAccession": func(accession string) string {
			return strings.ReplaceAll(accession, "-", "")
		},
		"AppVersion": func() string {
			return fmt.Sprintf("%s %s", s.SHA1Ver, s.BuildTime)
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
