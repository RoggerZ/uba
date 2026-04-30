package model

import "github.com/1340691923/xwl_bi/engine/db"

type ReportTable struct {
	Id         int    `db:"id" json:"id"`
	Appid      int    `db:"appid" json:"appid"`
	UserId     int    `db:"user_id" json:"user_id"`
	Name       string `db:"name" json:"name"`
	RtType     int8   `db:"rt_type" json:"rt_type"`
	Data       string `db:"data" json:"data"`
	CreateTime string `db:"create_time" json:"create_time"`
	UpdateTime string `db:"update_time" json:"update_time"`
	Remark     string `db:"remark" json:"remark"`
	PannelId   int    `db:"-" json:"pannel_id"`
	WindowSize string `db:"-" json:"window_size"`
}

func (this *ReportTable) InsertOrUpdate() (err error) {
	if this.Id > 0 {
		sql := `update report_table set name=?,rt_type=?,data=?,remark=? where id=? and appid=? and user_id=?`
		_, err = db.Sqlx.Exec(sql, this.Name, this.RtType, this.Data, this.Remark, this.Id, this.Appid, this.UserId)
		return
	}

	sql := `insert into report_table(appid,user_id,name,rt_type,data,remark) values(?,?,?,?,?,?)
			on duplicate key update id=LAST_INSERT_ID(id),data=values(data),remark=values(remark)`
	res, err := db.Sqlx.Exec(sql, this.Appid, this.UserId, this.Name, this.RtType, this.Data, this.Remark)
	if err != nil {
		return err
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	this.Id = int(lastInsertID)
	return nil
}
