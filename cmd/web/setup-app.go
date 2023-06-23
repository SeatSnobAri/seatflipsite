package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/SeatSnobAri/seatflipsite/internal/config"
	"github.com/SeatSnobAri/seatflipsite/internal/driver"
	"github.com/SeatSnobAri/seatflipsite/internal/handlers"
	"github.com/SeatSnobAri/seatflipsite/internal/helpers"
	"github.com/SeatSnobAri/seatflipsite/internal/render"
	"github.com/alexedwards/scs/postgresstore"
	"log"
	"net/http"
	"os"
	"time"
)

func setupApp() (*string, error) {
	// read flags
	insecurePort := flag.String("port", ":4000", "port to listen on")
	useCache := flag.Bool("cache", true, "Use template cache")
	identifier := flag.String("identifier", "seatflip", "unique identifier")
	domain := flag.String("domain", "localhost", "domain name (e.g. example.com)")
	inProduction := flag.Bool("production", false, "application is in production")
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbPort := flag.String("dbport", "5432", "database port")
	dbUser := flag.String("dbuser", "postgres", "database user")
	dbPass := flag.String("dbpass", "", "database password")
	databaseName := flag.String("db", "seatflip", "database name")
	dbSsl := flag.String("dbssl", "disable", "database ssl setting")
	pusherHost := flag.String("pusherHost", "", "pusher host")
	pusherPort := flag.String("pusherPort", "443", "pusher port")
	pusherApp := flag.String("pusherApp", "1618595", "pusher app id")
	pusherKey := flag.String("pusherKey", "ae30ade191f5e0f49b84", "pusher key")
	pusherSecret := flag.String("pusherSecret", "48c39a2ae19f52ac26f8", "pusher secret")
	pusherSecure := flag.Bool("pusherSecure", true, "pusher server uses SSL (true or false)")

	flag.Parse()

	if *dbUser == "" || *dbHost == "" || *dbPort == "" || *databaseName == "" || *identifier == "" {
		fmt.Println("Missing required flags.")
		os.Exit(1)
	}
	app.UseCache = *useCache
	log.Println("Connecting to database....")
	dsnString := ""

	// when developing locally, we often don't have a db password
	if *dbPass == "" {
		dsnString = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			*dbHost,
			*dbPort,
			*dbUser,
			*databaseName,
			*dbSsl)
	} else {
		dsnString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			*dbHost,
			*dbPort,
			*dbUser,
			*dbPass,
			*databaseName,
			*dbSsl)
	}

	db, err := driver.ConnectPostgres(dsnString)
	if err != nil {
		log.Fatal("Cannot connect to database!", err)
	}

	// session
	log.Printf("Initializing session manager....")
	session = scs.New()
	session.Store = postgresstore.New(db.SQL)
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.Name = fmt.Sprintf("gbsession_id_%s", *identifier)
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = *inProduction

	// define application configuration
	a := config.AppConfig{
		DB:           db,
		Session:      session,
		InProduction: *inProduction,
		Domain:       *domain,
		PusherSecret: *pusherSecret,
		Version:      seatflipVersion,
		Identifier:   *identifier,
	}

	app = a

	log.Println("Getting preferences...")
	preferenceMap = make(map[string]string)

	preferenceMap["pusher-host"] = *pusherHost
	preferenceMap["pusher-port"] = *pusherPort
	preferenceMap["pusher-key"] = *pusherKey
	preferenceMap["identifier"] = *identifier
	preferenceMap["version"] = seatflipVersion

	app.PreferenceMap = preferenceMap

	// create pusher client
	wsClient = pusher.Client{
		AppID:   *pusherApp,
		Key:     *pusherKey,
		Secret:  *pusherSecret,
		Cluster: "mt1",
		Secure:  true,
	}

	log.Println("Host", fmt.Sprintf("%s:%s", *pusherHost, *pusherPort))
	log.Println("Secure", *pusherSecure)

	app.WsClient = wsClient

	redis := redis.NewClient(&redis.Options{
		Addr:     ":6379",
		Password: "",
		DB:       0,
	})
	_, err = redis.Do(context.Background(), "CONFIG", "SET", "notify-keyspace-events", "KEA").Result() // this is telling redis to publish events since it's off by default.
	if err != nil {
		fmt.Printf("unable to set keyspace events %v", err.Error())
		os.Exit(1)
	}
	//pubsub := redis.PSubscribe(context.Background(), "__keyevent@0__:expired") // this is telling redis to subscribe to events published in the keyevent channel, specifically for expired events
	pubsub := redis.PSubscribe(context.Background(), "__keyevent@0__:expired") // this is telling redis to subscribe to events published in the keyevent channel, specifically for expired events
	app.Redis = redis

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Println(err)
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc

	repo = handlers.NewPostgresqlHandlers(db, &app)
	handlers.NewHandlers(repo, &app)
	render.NewRenderer(&app)

	helpers.NewHelpers(&app)

	go handlers.Repo.RedisExpiry(pubsub)

	return insecurePort, err
}

// createDirIfNotExist creates a directory if it does not exist
func createDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, mode)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
