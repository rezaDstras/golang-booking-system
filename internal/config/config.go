package config

import (
	"log"
	"text/template"

	"github.com/alexedwards/scs/v2"
	"github.com/rezaDastrs/internal/models"
)

//hold the application config

type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InProduction  bool
	Session       *scs.SessionManager
	InfoLog       *log.Logger
	ErrorLog      *log.Logger

	//access mail channel globally in project
	MailChan chan models.MailData
}
