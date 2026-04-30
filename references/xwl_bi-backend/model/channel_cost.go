package model

import (
	"github.com/1340691923/xwl_bi/engine/db"
)

type ChannelCost struct {
	ID        int64   `json:"id" db:"id"`
	AppID     int     `json:"appid" db:"app_id"`
	Channel   string  `json:"channel" db:"channel"`
	CostDate  string  `json:"costDate" db:"cost_date"` // Format: YYYY-MM-DD
	Cost      float64 `json:"cost" db:"cost"`
	CreatedAt string  `json:"createdAt" db:"create_time"`
	UpdatedAt string  `json:"updatedAt" db:"update_time"`
}

func (this *ChannelCost) TableName() string {
	return "channel_cost"
}

// Insert adds a new cost record
func (this *ChannelCost) Insert() (err error) {
	_, err = db.Sqlx.Exec("INSERT INTO channel_cost (app_id, channel, cost_date, cost) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE cost = VALUES(cost), update_time = CURRENT_TIMESTAMP",
		this.AppID, this.Channel, this.CostDate, this.Cost)
	return
}

// Update updates an existing cost record
func (this *ChannelCost) Update() (err error) {
	_, err = db.Sqlx.Exec("UPDATE channel_cost SET cost = ? WHERE id = ?",
		this.Cost, this.ID)
	return
}

// Delete removes a cost record
func (this *ChannelCost) Delete() (err error) {
	_, err = db.Sqlx.Exec("DELETE FROM channel_cost WHERE id = ?", this.ID)
	return
}

// List returns a list of cost records
func (this *ChannelCost) List(appid int, startDate, endDate string) (list []ChannelCost, err error) {
	err = db.Sqlx.Select(&list, "SELECT * FROM channel_cost WHERE app_id = ? AND cost_date BETWEEN ? AND ? ORDER BY cost_date DESC, channel ASC", appid, startDate, endDate)
	return
}

// GetByDateAndChannel returns a specific cost record
func (this *ChannelCost) GetByDateAndChannel(appid int, date, channel string) (cost ChannelCost, err error) {
	err = db.Sqlx.Get(&cost, "SELECT * FROM channel_cost WHERE app_id = ? AND cost_date = ? AND channel = ? LIMIT 1", appid, date, channel)
	return
}
