package server

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/sec"
	"github.com/gorilla/mux"
)

func (s Server) GenerateRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", s.HandlerHome).Methods("GET")
	router.HandleFunc("/fillings/{year}", s.HandlerMonthsPage).Methods("GET")
	router.HandleFunc("/fillings/{year}/{month}", s.HandlerDaysPage).Methods("GET")
	router.HandleFunc("/fillings/{year}/{month}/{day}", s.HandlerCompaniesPage).Methods("GET")
	router.HandleFunc("/fillings/{year}/{month}/{day}/{cik}", s.HandlerFillingsPage).Methods("GET")
	router.HandleFunc("/api/v1/uptime", s.HandlerUptime).Methods("GET")
	router.PathPrefix("/").HandlerFunc(s.HandlerFiles)

	return router
}

func (s Server) HandlerHome(w http.ResponseWriter, r *http.Request) {
	var err error

	content := make(map[string]interface{})
	content["Years"], err = sec.UniqueYearsInWorklist(s.DB)
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
	content["Months"], err = sec.MonthsInYearInWorklist(s.DB, year)
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

	days, err := secVar.GetFillingDaysFromMonthYear(s.DB, year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
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

	companies, err := secVar.GetFillingCompaniesFromYearMonthDay(s.DB, year, month, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["Day"] = day
	content["Companies"] = companies

	err = s.RenderTemplate(w, "companies.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) HandlerFillingsPage(w http.ResponseWriter, r *http.Request) {
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

	fillings, err := secVar.SearchFillingsByYearMonthDayCIK(s.DB, year, month, day, cik)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["Day"] = day
	content["CIK"] = cik
	content["Fillings"] = fillings

	err = s.RenderTemplate(w, "fillings.page.gohtml", content)
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
