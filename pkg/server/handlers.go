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
	"github.com/equres/sec/pkg/secworklist"
	"github.com/gorilla/mux"
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
	router.HandleFunc("/stats", s.HandlerStatsPage).Methods("GET")
	router.HandleFunc("/api/v1/uptime", s.HandlerUptime).Methods("GET")
	router.PathPrefix("/").HandlerFunc(s.HandlerFiles)
	return router, nil
}

func (s Server) HandlerHome(w http.ResponseWriter, r *http.Request) {
	content := make(map[string]interface{})
	recentFilings, err := sec.GetFiveRecentFilings(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type FormattedFiling struct {
		CompanyName string
		PubDate     string
		FormType    string
		XbrlURL     string
	}

	var recentFilingsFormatted []FormattedFiling

	for _, filing := range recentFilings {
		fileLink, err := url.Parse(filing.XbrlURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		formattedFiling := FormattedFiling{
			CompanyName: filing.CompanyName,
			PubDate:     filing.PubDate.Format("2006-01-02"),
			FormType:    filing.FormType,
			XbrlURL:     fileLink.Path,
		}
		recentFilingsFormatted = append(recentFilingsFormatted, formattedFiling)
	}
	content["RecentFilings"] = recentFilingsFormatted

	content["Years"], err = secworklist.UniqueYearsInWorklist(s.DB)
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
	content["Months"], err = secworklist.MonthsInYearInWorklist(s.DB, year)
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

	secVar, err := sec.NewSEC(s.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	days, err := secVar.GetFilingDaysFromMonthYear(s.DB, year, month)
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

	secVar, err := sec.NewSEC(s.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companies, err := secVar.GetFilingCompaniesFromYearMonthDay(s.DB, year, month, day)
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

	secVar, err := sec.NewSEC(s.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filings, err := secVar.SearchFilingsByYearMonthDayCIK(s.DB, year, month, day, cik)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	companyName, err := secVar.GetCompanyNameFromCIK(s.DB, cik)
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

func (s Server) HandlerStatsPage(w http.ResponseWriter, r *http.Request) {
	failedCount, err := sec.GetFailedDownloadEventCount(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	successCount, err := sec.GetSuccessfulDownloadEventCount(s.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	content := make(map[string]interface{})
	content["FailedDownloadsCount"] = failedCount
	content["SuccessfulDownloadsCount"] = successCount

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
