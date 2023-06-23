package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SeatSnobAri/seatflip2.0/internal/models"
)

func (repo *DBRepo) RedisExpiry(pubsub *redis.PubSub) {
	for { // infinite loop
		// this listens in the background for messages.
		message, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			fmt.Printf("error message - %v", err.Error())
			break
		}
		fmt.Printf("Keyspace event recieved %v  \n", message.String())

		var msg models.UflipPayload

		// get info
		row := repo.App.Redis.Get(context.Background(), message.Payload)
		json.Unmarshal([]byte(row.Val()), &msg)
		data := make(map[string]string)
		data["del"] = message.Payload

		repo.App.WsClient.Trigger("public-channel", "expired-row", data)
	}
}
