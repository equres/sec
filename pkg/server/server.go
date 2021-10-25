package server

import (
	"embed"
	"net/http"
	"time"

	"github.com/equres/sec/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
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

	logrus.Info("Listening on port", s.Config.Main.ServerPort)
	err := http.ListenAndServe(s.Config.Main.ServerPort, router)
	if err != nil {
		return err
	}
	return nil
}
