package nats

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"hezzl/logs"
	"hezzl/models"
	"log"
	"os"
)

var NC *nats.Conn

func Connect() error {
	nc, err := nats.Connect(fmt.Sprintf("nats://%s:4222", os.Getenv("NATS_HOST")))
	if err != nil {
		return err
	}

	NC = nc
	return nil
}

func Subscribe() error {
	_, err := NC.Subscribe(os.Getenv("NATS_QUEUE"), func(msg *nats.Msg) {
		itemJSON := string(msg.Data)
		var item models.ItemLog
		_ = json.Unmarshal([]byte(itemJSON), &item)

		_, err := logs.ClickhouseDB.Query(`INSERT INTO Items (Id, Name, CampaignId, Description, Priority, Removed, EventTime) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			item.ID,
			item.Name,
			item.CampaignID,
			item.Description,
			item.Priority,
			item.Removed,
			item.EventTime)
		if err != nil {
			log.Printf("Не удалось отправить лог: %s\n", err.Error())
		}
	})
	if err != nil {
		return err
	}
	return nil
}
