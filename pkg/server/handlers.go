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
	"time"

	"github.com/equres/sec/pkg/sec"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s Server) GenerateRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", s.HandlerHome).Methods("GET")
	router.HandleFunc("/static/{cik}/{accession}/{filename}", s.HandlerFiles).Methods("GET")
	router.HandleFunc("/search/{year}", s.HandlerMonthsPage).Methods("GET")
	router.HandleFunc("/search/{year}/{month}", s.HandlerFillingsPage).Methods("GET")

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
	vars := mux.Vars(r)
	cik, ok := vars["cik"]
	if !ok {
		fmt.Println("Cik number not there")
		http.Error(w, "404 File Not Found", http.StatusNotFound)
		return
	}

	accession, ok := vars["accession"]
	if !ok {
		fmt.Println("accession number not there")
		http.Error(w, "404 File Not Found", http.StatusNotFound)
		return
	}
	accession = strings.ReplaceAll(accession, "-", "")

	filename, ok := vars["filename"]
	if !ok {
		fmt.Println("filename not there")
		http.Error(w, "404 File Not Found", http.StatusNotFound)
		return
	}

	filePath := filepath.Join(s.Config.Main.CacheDir, "/Archives/edgar/data/", cik, accession, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = filepath.Join(s.Config.Main.CacheDirUnpacked, "/Archives/edgar/data/", cik, accession, filename)
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

	firstOfMonth, err := time.Parse("2006-1", fmt.Sprint(year, "-", month))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	secVar, err := sec.NewSEC(s.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fillings, err := secVar.SearchByFillingDate(s.DB, firstOfMonth, lastOfMonth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content := make(map[string]interface{})
	content["Year"] = year
	content["Month"] = month
	content["Fillings"] = fillings

	err = s.RenderTemplate(w, "fillings.page.gohtml", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	template, err := template.ParseFS(s.TemplatesFS, "templates/"+tmpl, "templates/base.layout.gohtml")
	if err != nil {
		return err
	}

	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		return err
	}

	return nil
}

func getIntVar(vars map[string]string, varName string) (int, error) {
	varStr, ok := vars[varName]
	if !ok {
		logrus.Error("please choose a proper year and month")
		return 0, errors.New("please choose a proper year and month")
	}
	value, err := strconv.Atoi(varStr)
	if err != nil {
		logrus.Error("please choose a proper year and month")
		return 0, errors.New("please choose a proper year and month")
	}
	return value, nil
}
