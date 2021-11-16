package server

import (
	"embed"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/config"
	"github.com/jmoiron/sqlx"
)

var (
	GlobalUptime    time.Time
	GlobalSHA1Ver   string // SHA1 revision used to build the program
	GlobalBuildTime string // when the executable was built
)

type Server struct {
	DB          *sqlx.DB
	Config      config.Config
	TemplatesFS embed.FS
	SHA1Ver     string
	BuildTime   string
}

func NewServer(db *sqlx.DB, config config.Config, templates embed.FS) (Server, error) {
	s := Server{
		DB:          db,
		Config:      config,
		TemplatesFS: templates,
		SHA1Ver:     GlobalSHA1Ver,
		BuildTime:   GlobalBuildTime,
	}

	return s, nil
}

func (s Server) StartServer() error {
	router := s.GenerateRouter()
	GlobalUptime = time.Now()

	log.Info(s.SHA1Ver)
	log.Info(s.BuildTime)
	log.Info("Listening on port", s.Config.Main.ServerPort)
	err := http.ListenAndServe(s.Config.Main.ServerPort, router)
	if err != nil {
		return err
	}
	return nil
}
