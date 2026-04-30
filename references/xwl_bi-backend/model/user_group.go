package model

import (
	"errors"
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

type UserGroup struct {
	Id                int      `db:"id" json:"id"`
	GroupName         string   `db:"group_name" json:"group_name"`
	GroupDisplayName  string   `db:"group_display_name" json:"group_display_name"`
	GroupRemark       string   `db:"group_remark" json:"group_remark"`
	UpdateType        string   `db:"update_type" json:"update_type"`
	CreateType        string   `db:"create_type" json:"create_type"`
	RuleContent       string   `db:"rule_content" json:"rule_content"`
	CreateBy          int      `db:"create_by" json:"create_by"`
	UserCount         int      `db:"user_count" json:"user_count"`
	UserList          []byte   `db:"user_list" json:"-"`
	UserListData      []string `db:"-" json:"user_list"`
	LastCalculateTime string   `db:"last_calculate_time" json:"last_calculate_time"`
	CanManualRefresh  bool     `db:"can_manual_refresh" json:"can_manual_refresh"`
	CreateTime        string   `db:"create_time" json:"create_time"`
	UpdateTime        string   `db:"update_time" json:"update_time"`
}

func duplicateUserGroupError(err error) error {
	if util.IsMysqlRepeatError(err) {
		return errors.New("分群名称或分群显示名重复，请重新填写")
	}
	return err
}

func (this *UserGroup) Insert(managerUid int32, appid int) (err error) {
	res, err := db.
		SqlBuilder.
		Insert("user_group").
		SetMap(map[string]interface{}{
			"group_name":          this.GroupName,
			"group_display_name":  this.GroupDisplayName,
			"group_remark":        this.GroupRemark,
			"update_type":         this.UpdateType,
			"create_type":         this.CreateType,
			"rule_content":        this.RuleContent,
			"create_by":           managerUid,
			"user_count":          this.UserCount,
			"appid":               appid,
			"user_list":           this.UserList,
			"last_calculate_time": this.LastCalculateTime,
			"can_manual_refresh":  this.CanManualRefresh,
		}).
		RunWith(db.Sqlx).
		Exec()
	if err != nil {
		return duplicateUserGroupError(err)
	}
	lastInsertID, err := res.LastInsertId()
	if err == nil {
		this.Id = int(lastInsertID)
	}
	return nil
}

func (this *UserGroup) Update(managerUid int32, appid int) (err error) {
	_, err = db.SqlBuilder.
		Update("user_group").
		SetMap(map[string]interface{}{
			"group_name":         this.GroupName,
			"group_display_name": this.GroupDisplayName,
			"group_remark":       this.GroupRemark,
			"update_type":        this.UpdateType,
			"create_type":        this.CreateType,
			"rule_content":       this.RuleContent,
			"user_count":         this.UserCount,
			"user_list":          this.UserList,
			"last_calculate_time": this.LastCalculateTime,
			"can_manual_refresh": this.CanManualRefresh,
		}).
		Where(
			db.Eq{
				"create_by": managerUid,
				"id":        this.Id,
				"appid":     appid,
			}).
		RunWith(db.Sqlx).
		Exec()
	if err != nil {
		return duplicateUserGroupError(err)
	}
	return nil
}

func (this *UserGroup) UpdateCalculatedData(managerUid int32, appid int) (err error) {
	_, err = db.SqlBuilder.
		Update("user_group").
		SetMap(map[string]interface{}{
			"user_count":          this.UserCount,
			"user_list":           this.UserList,
			"last_calculate_time": this.LastCalculateTime,
			"can_manual_refresh":  this.CanManualRefresh,
		}).
		Where(
			db.Eq{
				"create_by": managerUid,
				"id":        this.Id,
				"appid":     appid,
			}).
		RunWith(db.Sqlx).
		Exec()
	return
}

func (this *UserGroup) ModifyUserGroup(managerUid int32, appid int) (err error) {
	return this.Update(managerUid, appid)
}

func (this *UserGroup) DeleteUserGroupById(managerUid int32, appid int) (err error) {
	_, err = db.SqlBuilder.
		Delete("user_group").
		Where(db.Eq{"create_by": managerUid, "id": this.Id, "appid": appid}).
		RunWith(db.Sqlx).
		Exec()
	return
}

func (this *UserGroup) List(managerUid int32, appid int, keyword, updateType string) (list []UserGroup, err error) {
	builder := db.SqlBuilder.
		Select(
			"id",
			"group_name",
			"IFNULL(group_display_name, '') AS group_display_name",
			"group_remark",
			"IFNULL(update_type, 'manual') AS update_type",
			"IFNULL(create_type, 'analysis_page_snapshot') AS create_type",
			"IFNULL(rule_content, '') AS rule_content",
			"create_by",
			"user_count",
			"user_list",
			"IFNULL(last_calculate_time, create_time) AS last_calculate_time",
			"IFNULL(can_manual_refresh, 0) AS can_manual_refresh",
			"create_time",
			"update_time",
		).
		From("user_group").
		Where(db.Eq{"create_by": managerUid, "appid": appid})

	if updateType != "" {
		builder = builder.Where(db.Eq{"update_type": updateType})
	}

	if keyword != "" {
		builder = builder.Where(db.Or{
			db.Like{"group_name": "%" + keyword + "%"},
			db.Like{"group_display_name": "%" + keyword + "%"},
		})
	}

	SQL, args, err := builder.OrderBy("update_time desc", "id desc").ToSql()
	if err != nil {
		return nil, err
	}

	if err := db.Sqlx.Select(&list, SQL, args...); err != nil {
		return nil, err
	}
	return list, err
}

func (this *UserGroup) GetById(managerUid int32, appid int) (obj UserGroup, err error) {
	SQL, args, err := db.SqlBuilder.
		Select(
			"id",
			"group_name",
			"IFNULL(group_display_name, '') AS group_display_name",
			"group_remark",
			"IFNULL(update_type, 'manual') AS update_type",
			"IFNULL(create_type, 'analysis_page_snapshot') AS create_type",
			"IFNULL(rule_content, '') AS rule_content",
			"create_by",
			"user_count",
			"user_list",
			"IFNULL(last_calculate_time, create_time) AS last_calculate_time",
			"IFNULL(can_manual_refresh, 0) AS can_manual_refresh",
			"create_time",
			"update_time",
		).
		From("user_group").
		Where(db.Eq{"create_by": managerUid, "id": this.Id, "appid": appid}).
		ToSql()
	if err != nil {
		return obj, err
	}

	err = db.Sqlx.Get(&obj, SQL, args...)
	return obj, err
}

func (this *UserGroup) GetSelectOptions(managerUid int32, appid int) (list []UserGroup, err error) {
	SQL, args, err := db.SqlBuilder.
		Select(
			"id",
			"group_name",
			"IFNULL(group_display_name, '') AS group_display_name",
			"IFNULL(create_type, 'analysis_page_snapshot') AS create_type",
		).
		From("user_group").
		Where(
			db.Eq{
				"create_by": managerUid,
				"appid":     appid,
			}).
		OrderBy("update_time desc", "id desc").
		ToSql()
	if err != nil {
		return
	}

	err = db.Sqlx.Select(&list, SQL, args...)
	if err != nil {
		return
	}

	return

}
