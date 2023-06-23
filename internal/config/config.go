package config

import (
	"github.com/SeatSnobAri/seatflipsite/internal/driver"
	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go/v5"
	"github.com/redis/go-redis/v9"
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
