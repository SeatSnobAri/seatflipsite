package templates

import (
	"github.com/SeatSnobAri/seatflipsite/internal/forms"
	"github.com/SeatSnobAri/seatflipsite/internal/models"
)

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	PreferenceMap   map[string]string
	User            models.User
	Flash           string
	Warning         string
	Error           string
	Form            *forms.Form
	IsAuthenticated bool
}
