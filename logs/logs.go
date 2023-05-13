package logs

import (
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"os"
)

var ClickhouseDB *sql.DB

func Connect() error {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", os.Getenv("CLICKHOUSE_HOST"), 8123)},
		Auth: clickhouse.Auth{
			Database: os.Getenv("CLICKHOUSE_DB"),
			Username: os.Getenv("CLICKHOUSE_USER"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
		Protocol: clickhouse.HTTP,
	})
	ClickhouseDB = conn
	err := conn.Ping()
	if err != nil {
		return err
	}
	// я долго бился с migrate, но он не проводил миграции для clickhouse, выдавая bad connection
	// поэтому clickhouse мигрируем руками
	Migrate()
	return nil
}

func Migrate() {
	migrations := "CREATE TABLE Items (\n                       Id Int32,\n                       CampaignId Int32,\n                       Name String,\n                       Description Nullable(String),\n                       Priority Nullable(Int32),\n                       Removed Nullable(Bool),\n                       EventTime Nullable(DateTime)\n) ENGINE = MergeTree()\nORDER BY Id;"
	idInd := "ALTER TABLE default.Items ADD INDEX idx_items_id(Id) TYPE minmax GRANULARITY 8192;"
	campInd := "ALTER TABLE default.Items ADD INDEX idx_items_campaign_id(CampaignId) TYPE minmax GRANULARITY 8192;"
	nameInd := "ALTER TABLE default.Items ADD INDEX idx_items_name(Name) TYPE bloom_filter GRANULARITY 8192;"

	queries := []string{migrations, idInd, campInd, nameInd}
	for _, query := range queries {
		_, _ = ClickhouseDB.Exec(query)
	}
}
