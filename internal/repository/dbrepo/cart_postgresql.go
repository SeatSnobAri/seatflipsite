package dbrepo

import (
	"context"
	"github.com/SeatSnobAri/seatflipsite/internal/models"
	"strconv"
	"strings"
	"time"
)

func (repo *postgresDBRepo) InsertCart(payload models.UflipPayload, user models.UserPayload) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into "carts".carts (id, event_date, event_name, event_venue, seat_info, ticket_info, ticket_price,
                         ticket_total, bought, stock_type, user_id) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	tt := strings.ReplaceAll(payload.TicketTotal, "$", "")
	ticket_total, err := strconv.ParseFloat(tt, 64)
	if err != nil {
		return err
	}

	_, err = repo.DB.ExecContext(ctx, query,
		payload.UUID,
		payload.EventDate,
		payload.EventName,
		payload.EventVenue,
		payload.SeatInfo,
		payload.TicketInfo,
		payload.TicketPrice,
		ticket_total,
		payload.Buy,
		payload.StockType,
		user.Id,
	)
	if err != nil {
		return err
	}
	return nil

}

func (repo *postgresDBRepo) UpdateCart(buy bool, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `update "carts".carts set bought=$1 where id = $2`

	_, err := repo.DB.ExecContext(ctx, query, buy, id)
	if err != nil {
		return err
	}
	return nil

}

func (repo *postgresDBRepo) GetCartUser(id string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `select user_id from "carts".carts where id=$1`
	var userId string
	row := repo.DB.QueryRowContext(ctx, query, id)
	row.Scan(&userId)
	return userId

}
