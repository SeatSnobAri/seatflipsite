package handlers

import (
	"fmt"
	"github.com/pusher/pusher-http-go/v5"
	"io"
	"log"
	"net/http"
)

func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	userID := repo.App.Session.Get(r.Context(), "user_id")

	u, _ := repo.DB.GetUserById(userID.(string))

	params, _ := io.ReadAll(r.Body)

	presenceData := pusher.MemberData{
		UserID: userID.(string),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   userID.(string),
		},
	}

	response, err := app.WsClient.AuthorizePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response)
}

// SendPrivateMessage is sample code for sending to private channel
func (repo *DBRepo) SendPrivateMessage(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	id := r.URL.Query().Get("id")

	data := make(map[string]string)
	data["message"] = msg

	_ = repo.App.WsClient.Trigger(fmt.Sprintf("private-channel-%s", id), "private-message", data)

}

func (repo *DBRepo) broadcastMessage(channel, messageType string, data map[string]string) {
	err := app.WsClient.Trigger(channel, messageType, data)
	if err != nil {
		log.Println(err)
	}
}
