package server

import (
	"embed"
	"log"
	"net/http"
	"time"

	"github.com/equres/sec/pkg/config"
	"github.com/jmoiron/sqlx"
)

var GlobalUptime time.Time

type Server struct {
	DB          *sqlx.DB
	Config      config.Config
	TemplatesFS embed.FS
}

func NewServer(db *sqlx.DB, config config.Config, templates embed.FS) (Server, error) {
	s := Server{
		DB:          db,
		Config:      config,
		TemplatesFS: templates,
	}

	return s, nil
}

func (s Server) StartServer() error {
	router := s.GenerateRouter()
	GlobalUptime = time.Now()

	log.Println("Listening on port", s.Config.Main.ServerPort)
	err := http.ListenAndServe(s.Config.Main.ServerPort, router)
	if err != nil {
		return err
	}
	return nil
}
