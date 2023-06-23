package config

import (
	"github.com/SeatSnobAri/seatflipsite/internal/driver"
	"html/template"
)

// AppConfig holds application configuration
type AppConfig struct {
	UseCache      bool
	DB            *driver.DB
	Session       *scs.SessionManager
	InProduction  bool
	Domain        string
	PreferenceMap map[string]string
	Redis         *redis.Client
	WsClient      pusher.Client
	PusherSecret  string
	TemplateCache map[string]*template.Template
	Version       string
	Identifier    string
}
