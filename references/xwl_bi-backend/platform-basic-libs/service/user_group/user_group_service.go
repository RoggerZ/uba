package user_group

import (
	"errors"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"strings"
	"time"
)

const (
	UpdateTypeManual             = "manual"
	CreateTypeUserGroupPageRule  = "user_group_page_rule"
	CreateTypeAnalysisPageRule   = "analysis_page_rule"
	CreateTypeAnalysisPageStatic = "analysis_page_snapshot"
)

type UserGroupService struct {
	ManagerID int32
	Appid     int
}

func gzipUserIDs(uids []string) ([]byte, error) {
	return util.GzipCompress(strings.Join(uids, ","))
}

func unzipUserIDs(data []byte) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	userListData, err := util.GzipUnCompress(data)
	if err != nil {
		return nil, err
	}
	if userListData == "" {
		return []string{}, nil
	}
	return strings.Split(userListData, ","), nil
}

func normalizeUniqueUserIDs(uids []string) []string {
	seen := make(map[string]struct{}, len(uids))
	res := make([]string, 0, len(uids))
	for _, uid := range uids {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		res = append(res, uid)
	}
	return res
}

func (this *UserGroupService) nowString() string {
	return time.Now().Format(util.TimeFormat)
}

func (this *UserGroupService) AddUserGroup(userCount int, uids []string, groupRemark, groupName string) (err error) {
	uids = normalizeUniqueUserIDs(uids)
	b, err := gzipUserIDs(uids)
	if err != nil {
		return
	}
	userGroup := model.UserGroup{
		GroupRemark:       groupRemark,
		GroupName:         groupName,
		GroupDisplayName:  groupName,
		UpdateType:        UpdateTypeManual,
		CreateType:        CreateTypeAnalysisPageStatic,
		RuleContent:       "",
		UserCount:         userCount,
		UserList:          b,
		LastCalculateTime: this.nowString(),
		CanManualRefresh:  false,
	}
	return userGroup.Insert(this.ManagerID, this.Appid)
}

func (this *UserGroupService) ModifyUserGroup(id int, groupRemark, groupName string) (err error) {
	current := model.UserGroup{Id: id}
	current, err = current.GetById(this.ManagerID, this.Appid)
	if err != nil {
		return err
	}
	current.GroupName = groupName
	if strings.TrimSpace(current.GroupDisplayName) == "" {
		current.GroupDisplayName = groupName
	}
	current.GroupRemark = groupRemark
	return current.Update(this.ManagerID, this.Appid)
}

func (this *UserGroupService) DeleteUserGroup(id int) (err error) {
	userGroup := model.UserGroup{}
	userGroup.Id = id
	return userGroup.DeleteUserGroupById(this.ManagerID, this.Appid)
}

func (this *UserGroupService) UserGroupList(keyword, updateType string) (list []model.UserGroup, err error) {
	userGroup := model.UserGroup{}
	list, err = userGroup.List(this.ManagerID, this.Appid, keyword, updateType)
	if err != nil {
		return nil, err
	}

	for index := range list {
		userListData, err := unzipUserIDs(list[index].UserList)
		if err != nil {
			return nil, err
		}
		list[index].UserListData = userListData
	}

	return list, err
}

func (this *UserGroupService) Detail(id int) (obj model.UserGroup, err error) {
	userGroup := model.UserGroup{Id: id}
	obj, err = userGroup.GetById(this.ManagerID, this.Appid)
	if err != nil {
		return obj, err
	}
	obj.UserListData, err = unzipUserIDs(obj.UserList)
	return obj, err
}

func (this *UserGroupService) buildUserGroupModel(req request.SaveUserGroupReq, userIDs []string, canManualRefresh bool) (model.UserGroup, error) {
	userIDs = normalizeUniqueUserIDs(userIDs)
	b, err := gzipUserIDs(userIDs)
	if err != nil {
		return model.UserGroup{}, err
	}

	updateType := strings.TrimSpace(req.UpdateType)
	if updateType == "" {
		updateType = UpdateTypeManual
	}

	createType := strings.TrimSpace(req.CreateType)
	if createType == "" {
		if len(req.RuleContent) > 0 {
			createType = CreateTypeUserGroupPageRule
		} else {
			createType = CreateTypeAnalysisPageStatic
		}
	}

	return model.UserGroup{
		Id:                req.Id,
		GroupName:         strings.TrimSpace(req.GroupName),
		GroupDisplayName:  strings.TrimSpace(req.GroupDisplayName),
		GroupRemark:       strings.TrimSpace(req.Remark),
		UpdateType:        updateType,
		CreateType:        createType,
		RuleContent:       strings.TrimSpace(string(req.RuleContent)),
		UserCount:         len(userIDs),
		UserList:          b,
		LastCalculateTime: this.nowString(),
		CanManualRefresh:  canManualRefresh,
	}, nil
}

func (this *UserGroupService) SaveUserGroup(req request.SaveUserGroupReq) (obj model.UserGroup, err error) {
	if strings.TrimSpace(req.GroupName) == "" {
		return obj, errors.New("分群名称不能为空")
	}
	if strings.TrimSpace(req.GroupDisplayName) == "" {
		return obj, errors.New("分群显示名不能为空")
	}

	var current model.UserGroup
	if req.Id > 0 {
		current, err = this.Detail(req.Id)
		if err != nil {
			return obj, err
		}
	}

	ruleContent := strings.TrimSpace(string(req.RuleContent))
	createType := strings.TrimSpace(req.CreateType)
	if createType == "" {
		createType = CreateTypeUserGroupPageRule
	}
	isRuleGroup := createType != CreateTypeAnalysisPageStatic
	var userIDs []string
	canManualRefresh := false

	if isRuleGroup {
		if ruleContent == "" {
			return obj, errors.New("规则型分群缺少规则内容")
		}
		userIDs, err = this.EvaluateRuleContent(ruleContent)
		if err != nil {
			return obj, err
		}
		canManualRefresh = true
	} else {
		userIDs = req.SnapshotUserList
		if len(userIDs) == 0 && ruleContent != "" {
			userIDs, err = this.EvaluateRuleContent(ruleContent)
			if err != nil {
				return obj, err
			}
		}
		if len(userIDs) == 0 && req.Id > 0 {
			userIDs = current.UserListData
		}
		canManualRefresh = false
	}

	groupModel, err := this.buildUserGroupModel(req, userIDs, canManualRefresh)
	if err != nil {
		return obj, err
	}
	if !isRuleGroup {
		groupModel.RuleContent = ""
	}

	if req.Id > 0 {
		if groupModel.CreateType == "" {
			groupModel.CreateType = current.CreateType
		}
		if groupModel.UpdateType == "" {
			groupModel.UpdateType = current.UpdateType
		}
		err = groupModel.Update(this.ManagerID, this.Appid)
	} else {
		err = groupModel.Insert(this.ManagerID, this.Appid)
	}
	if err != nil {
		return obj, err
	}

	return this.Detail(groupModel.Id)
}

func (this *UserGroupService) RefreshUserGroup(id int) (obj model.UserGroup, err error) {
	current, err := this.Detail(id)
	if err != nil {
		return obj, err
	}
	if !current.CanManualRefresh || strings.TrimSpace(current.RuleContent) == "" {
		return obj, errors.New("当前分群不支持手动更新")
	}

	userIDs, err := this.EvaluateRuleContent(current.RuleContent)
	if err != nil {
		return obj, err
	}
	userIDs = normalizeUniqueUserIDs(userIDs)
	b, err := gzipUserIDs(userIDs)
	if err != nil {
		return obj, err
	}

	current.UserCount = len(userIDs)
	current.UserList = b
	current.LastCalculateTime = this.nowString()
	current.CanManualRefresh = true

	err = current.UpdateCalculatedData(this.ManagerID, this.Appid)
	if err != nil {
		return obj, err
	}

	return this.Detail(id)
}

func (this *UserGroupService) Options() (list []model.UserGroup, err error) {
	userGroup := model.UserGroup{}
	list, err = userGroup.GetSelectOptions(this.ManagerID, this.Appid)
	return
}
