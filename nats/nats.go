package nats

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"hezzl/logs"
	"hezzl/models"
	"os"
)

var NC *nats.Conn

func Connect() {
	nc, _ := nats.Connect(fmt.Sprintf("nats://%s:4222", os.Getenv("NATS_HOST")))
	NC = nc
}

func Subscribe() {
	_, _ = NC.Subscribe(os.Getenv("NATS_QUEUE"), func(msg *nats.Msg) {
		itemJSON := string(msg.Data)
		var item models.ItemLog
		_ = json.Unmarshal([]byte(itemJSON), &item)

		logs.ClickhouseDB.Query(`INSERT INTO Items (Id, Name, CampaignId, Description, Priority, Removed, EventTime) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			item.ID,
			item.Name,
			item.CampaignID,
			item.Description,
			item.Priority,
			item.Removed,
			item.EventTime)
	})
}
