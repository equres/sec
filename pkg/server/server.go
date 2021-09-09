package server

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/sec"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	TemplateCache map[string]*template.Template
	DB            *sqlx.DB
	Config        config.Config
}

func NewServer() (Server, error) {
	s := Server{}

	templateCache, err := GenerateTemplateCache()
	if err != nil {
		return s, err
	}

	s.TemplateCache = templateCache

	return s, nil
}

var functions = template.FuncMap{
	"CreateRoute": func(names ...string) string {
		var result string
		for _, v := range names {
			result = fmt.Sprint(result, "/", v)
		}
		return result
	},
}

func (s Server) StartServer() error {
	router := mux.NewRouter()
	router.HandleFunc("/", s.Home).Methods("GET")
	router.HandleFunc("/{year}", s.MonthsPage).Methods("GET")
	router.HandleFunc("/{year}/{month}", s.FillingsPage).Methods("GET")
	err := http.ListenAndServe(":8000", router)
	if err != nil {
		return err
	}
	return nil
}

func (s Server) Home(w http.ResponseWriter, r *http.Request) {
	worklist, err := sec.WorklistWillDownloadGet(s.DB)
	if err != nil {
		return
	}

	content := make(map[string]interface{})
	yearsMap := make(map[int]bool)
	for _, v := range worklist {
		yearsMap[v.Year] = true
	}

	yearsSlice := []int{}
	for k := range yearsMap {
		yearsSlice = append(yearsSlice, k)
	}

	content["Years"] = yearsSlice

	err = s.RenderTemplate(w, "index.page.gohtml", content)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	return
}

func (s Server) MonthsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	yearArg, ok := vars["year"]
	if !ok {
		fmt.Fprint(w, fmt.Errorf("please choose a proper year"))
		return
	}
	year, err := strconv.Atoi(yearArg)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	worklist, err := sec.WorklistWillDownloadGet(s.DB)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	content := make(map[string]interface{})
	months := []int{}
	for _, v := range worklist {
		if v.Year == year {
			months = append(months, v.Month)
		}
	}
	content["Year"] = yearArg
	content["Months"] = months

	err = s.RenderTemplate(w, "months.page.gohtml", content)
	if err != nil {
		fmt.Fprint(w, err)
	}
	return
}

func (s Server) FillingsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	yearArg, ok := vars["year"]
	if !ok {
		fmt.Fprint(w, fmt.Errorf("please choose a proper year"))
		return
	}
	year, err := strconv.Atoi(yearArg)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	monthArg, ok := vars["month"]
	if !ok {
		fmt.Fprint(w, fmt.Errorf("please choose a proper month"))
		return
	}

	month, err := strconv.Atoi(monthArg)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	firstOfMonth, err := time.Parse("2006-1", fmt.Sprint(year, "-", month))
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	secVar, err := sec.NewSEC(s.Config)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	fillings, err := secVar.SearchByFillingDate(s.DB, firstOfMonth.Format("2006-01-02"), lastOfMonth.Format("2006-01-02"))
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	content := make(map[string]interface{})
	content["Year"] = yearArg
	content["Month"] = monthArg
	content["Fillings"] = fillings

	err = s.RenderTemplate(w, "fillings.page.gohtml", content)
	if err != nil {
		fmt.Fprint(w, err)
	}
	return
}

func (s Server) RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	template, ok := s.TemplateCache[tmpl]
	if !ok {
		return errors.New("could not find file: " + tmpl)
	}

	buf := new(bytes.Buffer)

	err := template.Execute(buf, data)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}
	return nil
}

func GenerateTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.gohtml")
	if err != nil {
		return cache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return cache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.gohtml")
		if err != nil {
			return cache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.gohtml")
			if err != nil {
				return cache, err
			}
		}

		cache[name] = ts
	}

	return cache, nil
}
