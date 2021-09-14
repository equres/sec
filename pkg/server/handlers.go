package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/sec"
	"github.com/gorilla/mux"
)

func (s Server) GenerateRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", s.HandlerHome).Methods("GET")
	router.HandleFunc("/{year}", s.HandlerMonthsPage).Methods("GET")
	router.HandleFunc("/{year}/{month}", s.HandlerFillingsPage).Methods("GET")

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
		return 0, fmt.Errorf("please choose a proper year and month")
	}
	value, err := strconv.Atoi(varStr)
	if err != nil {
		return 0, fmt.Errorf("please choose a proper year and month")
	}
	return value, nil
}