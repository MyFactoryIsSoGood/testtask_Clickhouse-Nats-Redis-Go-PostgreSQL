package models

import "time"

type Item struct {
	ID          int       `json:"id"`
	CampaignID  int       `json:"campaignId"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ItemLog struct {
	ID          int       `json:"id"`
	CampaignID  int       `json:"campaignId"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	EventTime   time.Time `json:"eventTime"`
}
