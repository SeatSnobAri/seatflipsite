package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SeatSnobAri/seatflip2.0/internal/config"
	"github.com/SeatSnobAri/seatflip2.0/internal/models"
	"github.com/SeatSnobAri/seatflip2.0/internal/templates"
	"github.com/justinas/nosurf"
	"io"

	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var app *config.AppConfig
var src = rand.NewSource(time.Now().UnixNano())

// NewHelpers creates new helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

// IsAuthenticated returns true if a user is authenticated
func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}

// RandomString returns a random string of letters of length n
func RandomString(n int) string {
	b := make([]byte, n)

	for i, theCache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			theCache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(theCache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		theCache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// ServerError will display error page for internal server error
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = log.Output(2, trace)

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	http.ServeFile(w, r, "./ui/static/500.html")
}

// DefaultData adds default data which is accessible to all templates
func DefaultData(td templates.TemplateData, r *http.Request, w http.ResponseWriter) templates.TemplateData {
	td.CSRFToken = nosurf.Token(r)
	td.IsAuthenticated = IsAuthenticated(r)
	td.PreferenceMap = app.PreferenceMap
	// if logged in, store user id in template data
	if td.IsAuthenticated {
		u := app.Session.Get(r.Context(), "user").(models.User)
		td.User = u
	}

	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")

	return td
}

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ReadJSON tries to read the body of a request and converts it into JSON
func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

// writeJSON takes a response status code and arbitrary data and writes a json response to the client
func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// errorJSON takes an error, and optionally a response status code, and generates and sends
// a json error response
func ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return WriteJSON(w, statusCode, payload)
}
