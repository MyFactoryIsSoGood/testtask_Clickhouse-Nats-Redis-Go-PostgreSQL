package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"hezzl/cache"
	"hezzl/db"
	"hezzl/logs"
	"hezzl/models"
	"hezzl/nats"
	"net/http"
	"os"
	"strconv"
	"time"
)

const ErrorItemNotFound = "errors.item.notFound"

func CreateItem(c *gin.Context) {
	companyId, _ := strconv.Atoi(c.Query("campaignId"))
	payload := models.Item{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	item := models.Item{
		CampaignID:  companyId,
		Name:        payload.Name,
		Description: payload.Description,
	}

	rows, err := db.DB_conn.Query(`SELECT * FROM campaigns WHERE id = $1`, item.CampaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if !rows.Next() { //если строк нет
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": ErrorItemNotFound, "details": []string{}})
		return
	}

	//cоздаем запись с приоритетом MAX(priority)+1, если строк в таблице нет, то priority = 1
	_, err = db.DB_conn.Query(`INSERT INTO items (name, campaign_id, description, priority) VALUES ($1,$2,$3, (SELECT COALESCE(MAX(priority), 0) + 1 FROM items))`,
		item.Name,
		item.CampaignID,
		item.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	//получаем созданную запись. Во избежание получения более старших дубликатов - сортируем по дате создания и берем самый новый
	rows, err = db.DB_conn.Query(`SELECT * FROM items WHERE name = $1 AND campaign_id = $2 AND description = $3 ORDER BY created_at DESC LIMIT 1`,
		item.Name,
		item.CampaignID,
		item.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	rows.Next() //готовим строку к биндингу
	resp := models.Item{}
	err = rows.Scan(&resp.ID, &resp.CampaignID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
	_ = rows.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	//Инвалидируем кэш
	_ = cache.InvalidateItems()

	itemLog := models.ItemLog{
		ID:          resp.ID,
		CampaignID:  resp.CampaignID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     resp.Removed,
		EventTime:   time.Now(),
	}
	itemJSON, _ := json.Marshal(itemLog)
	_ = nats.NC.Publish(os.Getenv("NATS_QUEUE"), itemJSON)
	c.JSON(http.StatusOK, resp)
	return
}

func GetItems(c *gin.Context) {
	var resp []models.Item
	found, items := cache.GetItems("items")
	if found {
		resp = items
		c.JSON(http.StatusOK, resp)
		return
	}

	rows, err := db.DB_conn.Query("SELECT * FROM items")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	for rows.Next() {
		item := models.Item{}
		_ = rows.Scan(&item.ID, &item.CampaignID, &item.Name, &item.Description, &item.Priority, &item.Removed, &item.CreatedAt)
		resp = append(resp, item)
	}
	_ = rows.Close()

	if resp == nil {
		c.JSON(http.StatusOK, []models.Item{})
		return
	}
	cache.SetItems("items", resp)

	c.JSON(http.StatusOK, resp)
}

// UpdateItem обновляет запись. Исходя из т.з. предположил, что изменять можно только name и description, т.к. removed
// выставляется отдельным методом, priority выставяляется системно, а created и campaign_id не стоит изменять.
func UpdateItem(c *gin.Context) {
	itemID, _ := strconv.Atoi(c.Query("id"))
	campaignID, _ := strconv.Atoi(c.Query("campaignId"))

	transaction, err := db.DB_conn.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	rows, err := transaction.Query(`SELECT * FROM items WHERE id = $1 AND campaign_id = $2 FOR UPDATE`, itemID, campaignID)
	if err != nil {
		_ = transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if !rows.Next() {
		_ = transaction.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": ErrorItemNotFound, "details": []string{}})
		return
	}
	_ = rows.Close()

	//Для разделения поведения при отсутствии поля description и при наличии description=""
	payload := models.Item{Description: "not provided"}

	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = transaction.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	if payload.Name == "" {
		_ = transaction.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"message": "errors.item.emptyName"})
		return
	}

	var query string
	if payload.Description == "not provided" { // если поле не пришло, оставляем описание неизменным
		query = "UPDATE items SET name = $1 WHERE id = $2 AND campaign_id = $3"
		_, err = transaction.Exec(query,
			payload.Name,
			itemID,
			campaignID)
	} else { // если поле пришло, заменяем описание. Даже на пустую строку
		query = "UPDATE items SET name = $1, description = $2 WHERE id = $3 AND campaign_id = $4"
		_, err = transaction.Exec(query,
			payload.Name,
			payload.Description,
			itemID,
			campaignID)
	}

	if err != nil {
		_ = transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	rows, err = transaction.Query(`SELECT * FROM items WHERE id = $1 AND campaign_id = $2`, itemID, campaignID)

	if err != nil {
		_ = transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if !rows.Next() { // If no rows found
		_ = transaction.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": ErrorItemNotFound, "details": []string{}})
		return
	}

	resp := models.Item{}
	err = rows.Scan(&resp.ID, &resp.CampaignID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
	_ = rows.Close()

	if err != nil {
		_ = transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	err = transaction.Commit()
	if err != nil {
		_ = transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	//Инвалидируем кэш
	_ = cache.InvalidateItems()

	itemLog := models.ItemLog{
		ID:          resp.ID,
		CampaignID:  resp.CampaignID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     resp.Removed,
		//Задавать время на стороне БД с помощью NOW() - неправильно. При недоступе базы время будет искажено
		EventTime: time.Now(),
	}
	itemJSON, _ := json.Marshal(itemLog)
	_ = nats.NC.Publish(os.Getenv("NATS_QUEUE"), itemJSON)
	c.JSON(http.StatusOK, resp)
}

func DeleteItem(c *gin.Context) {
	itemID, _ := strconv.Atoi(c.Query("id"))
	campaignID, _ := strconv.Atoi(c.Query("campaignId"))
	resp := models.Item{}

	rows, err := db.DB_conn.Query(`SELECT * FROM items WHERE id = $1 AND campaign_id = $2`, itemID, campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": ErrorItemNotFound, "details": []string{}})
		return
	} else {
		err = rows.Scan(&resp.ID, &resp.CampaignID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
		_ = rows.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}
	_ = rows.Close()

	_, err = db.DB_conn.Query(`UPDATE items SET removed = true WHERE id = $1 AND campaign_id = $2`,
		itemID,
		campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	resp.Removed = true
	itemLog := models.ItemLog{
		ID:          resp.ID,
		CampaignID:  resp.CampaignID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     resp.Removed,
		//Задавать время на стороне БД с помощью NOW() - неправильно. При недоступе базы время будет искажено
		EventTime: time.Now(),
	}
	itemJSON, _ := json.Marshal(itemLog)
	_ = nats.NC.Publish(os.Getenv("NATS_QUEUE"), itemJSON)
	_ = cache.InvalidateItems()

	c.JSON(http.StatusOK, gin.H{"id": resp.ID, "campaignId": resp.CampaignID, "removed": resp.Removed})
}

// GetLogs - обработчик для получения логов из Clickhouse, чтобы их можно было легко проверить
func GetLogs(c *gin.Context) {
	rows, _ := logs.ClickhouseDB.Query("SELECT Id,CampaignId,Name,CAST(Description as VARCHAR),CAST(Priority as INT),CAST(Removed as BOOL),CAST(EventTime as timestamp) FROM Items ORDER BY EventTime DESC\n")
	defer rows.Close()

	var logItems []models.ItemLog
	for rows.Next() {
		var log models.ItemLog
		err := rows.Scan(&log.ID, &log.CampaignID, &log.Name, &log.Description, &log.Priority, &log.Removed, &log.EventTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logItems = append(logItems, log)
	}

	c.JSON(http.StatusOK, logItems)
}
