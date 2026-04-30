package controller

import (
	"errors"
	"github.com/1340691923/xwl_bi/platform-basic-libs/jwt"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/response"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/user_group"
	"github.com/gofiber/fiber/v2"
	"strings"
)

type UserGroupController struct {
	BaseController
}

func (this UserGroupController) buildService(appid int, token string) user_group.UserGroupService {
	c, _ := jwt.ParseToken(token)
	return user_group.UserGroupService{
		ManagerID: c.UserID,
		Appid:     appid,
	}
}

//新增用户分群
func (this UserGroupController) AddUserGroup(ctx *fiber.Ctx) error {
	var addUserGroup request.AddUserGroup
	if err := ctx.BodyParser(&addUserGroup); err != nil {
		return this.Error(ctx, err)
	}

	if strings.TrimSpace(addUserGroup.Name) == "" {
		return this.Error(ctx, errors.New("用户分群名称不能为空"))
	}
	if len(addUserGroup.Ids) == 0 {
		return this.Error(ctx, errors.New("待分群用户ID列表不能为空"))
	}

	userGroupService := this.buildService(addUserGroup.Appid, this.GetToken(ctx))

	err := userGroupService.AddUserGroup(len(addUserGroup.Ids), addUserGroup.Ids, addUserGroup.Remark, addUserGroup.Name)
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.OperateSuccess, nil)
}

//修改用户分群
func (this UserGroupController) ModifyUserGroup(ctx *fiber.Ctx) error {
	var modifyUserGroup request.ModifyUserGroup
	if err := ctx.BodyParser(&modifyUserGroup); err != nil {
		return this.Error(ctx, err)
	}

	if strings.TrimSpace(modifyUserGroup.Name) == "" {
		return this.Error(ctx, errors.New("用户分群名称不能为空"))
	}

	if modifyUserGroup.Id == 0 {
		return this.Error(ctx, errors.New("用户分群ID不能为空"))
	}

	userGroupService := this.buildService(modifyUserGroup.Appid, this.GetToken(ctx))

	err := userGroupService.ModifyUserGroup(modifyUserGroup.Id, modifyUserGroup.Remark, modifyUserGroup.Name)
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.OperateSuccess, nil)
}

//保存用户分群（规则型/静态型）
func (this UserGroupController) SaveUserGroup(ctx *fiber.Ctx) error {
	var saveReq request.SaveUserGroupReq
	if err := ctx.BodyParser(&saveReq); err != nil {
		return this.Error(ctx, err)
	}

	if strings.TrimSpace(saveReq.GroupName) == "" {
		return this.Error(ctx, errors.New("分群名称不能为空"))
	}
	if strings.TrimSpace(saveReq.GroupDisplayName) == "" {
		return this.Error(ctx, errors.New("分群显示名不能为空"))
	}

	userGroupService := this.buildService(saveReq.Appid, this.GetToken(ctx))
	res, err := userGroupService.SaveUserGroup(saveReq)
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.OperateSuccess, res)
}

//手动更新用户分群
func (this UserGroupController) RefreshUserGroup(ctx *fiber.Ctx) error {
	var refreshReq request.RefreshUserGroupReq
	if err := ctx.BodyParser(&refreshReq); err != nil {
		return this.Error(ctx, err)
	}

	if refreshReq.Id == 0 {
		return this.Error(ctx, errors.New("用户分群ID不能为空"))
	}

	userGroupService := this.buildService(refreshReq.Appid, this.GetToken(ctx))
	res, err := userGroupService.RefreshUserGroup(refreshReq.Id)
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.OperateSuccess, res)
}

//删除用户分群
func (this UserGroupController) DeleteUserGroup(ctx *fiber.Ctx) error {
	var deleteUserGroup request.DeleteUserGroup
	if err := ctx.BodyParser(&deleteUserGroup); err != nil {
		return this.Error(ctx, err)
	}

	if deleteUserGroup.Id == 0 {
		return this.Error(ctx, errors.New("用户分群ID不能为空"))
	}

	userGroupService := this.buildService(deleteUserGroup.Appid, this.GetToken(ctx))

	err := userGroupService.DeleteUserGroup(deleteUserGroup.Id)
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.OperateSuccess, nil)
}

//用户分群列表
func (this UserGroupController) UserGroupList(ctx *fiber.Ctx) error {
	var userGroupList request.UserGroupListReq
	if err := ctx.BodyParser(&userGroupList); err != nil {
		return this.Error(ctx, err)
	}

	userGroupService := this.buildService(userGroupList.Appid, this.GetToken(ctx))

	list, err := userGroupService.UserGroupList(strings.TrimSpace(userGroupList.Keyword), strings.TrimSpace(userGroupList.UpdateType))
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.SearchSuccess, list)
}

//用户分群详情
func (this UserGroupController) UserGroupDetail(ctx *fiber.Ctx) error {
	var detailReq request.UserGroupDetailReq
	if err := ctx.BodyParser(&detailReq); err != nil {
		return this.Error(ctx, err)
	}
	if detailReq.Id == 0 {
		return this.Error(ctx, errors.New("用户分群ID不能为空"))
	}

	userGroupService := this.buildService(detailReq.Appid, this.GetToken(ctx))
	res, err := userGroupService.Detail(detailReq.Id)
	if err != nil {
		return this.Error(ctx, err)
	}
	return this.Success(ctx, response.SearchSuccess, res)
}

//用户分群下拉选
func (this UserGroupController) UserGroupSelect(ctx *fiber.Ctx) error {
	var userGroupList request.UserGroupList
	if err := ctx.BodyParser(&userGroupList); err != nil {
		return this.Error(ctx, err)
	}

	userGroupService := this.buildService(userGroupList.Appid, this.GetToken(ctx))

	list, err := userGroupService.Options()
	if err != nil {
		return this.Error(ctx, err)
	}

	return this.Success(ctx, response.SearchSuccess, list)
}
