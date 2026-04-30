package cost

import (
	"errors"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/gofiber/fiber/v2"
	"time"
)

type ChannelCostService struct {
	Ctx *fiber.Ctx
}

func (s *ChannelCostService) Add(req request.ChannelCostAddReq) error {
	if req.AppID == 0 || req.Channel == "" || req.CostDate == "" {
		return errors.New("missing required fields")
	}

	// Verify date format
	_, err := time.Parse("2006-01-02", req.CostDate)
	if err != nil {
		return errors.New("invalid date format, expected YYYY-MM-DD")
	}

	cost := model.ChannelCost{
		AppID:    req.AppID,
		Channel:  req.Channel,
		CostDate: req.CostDate,
		Cost:     req.Cost,
	}

	return cost.Insert()
}

func (s *ChannelCostService) Update(req request.ChannelCostUpdateReq) error {
	if req.ID == 0 {
		return errors.New("missing id")
	}
	cost := model.ChannelCost{
		ID:   req.ID,
		Cost: req.Cost,
	}
	return cost.Update()
}

func (s *ChannelCostService) Delete(req request.ChannelCostDeleteReq) error {
	if req.ID == 0 {
		return errors.New("missing id")
	}
	cost := model.ChannelCost{
		ID: req.ID,
	}
	return cost.Delete()
}

func (s *ChannelCostService) List(req request.ChannelCostListReq) ([]model.ChannelCost, error) {
	if req.AppID == 0 {
		return nil, errors.New("missing appid")
	}
	if req.StartDate == "" || req.EndDate == "" {
		return nil, errors.New("missing date range")
	}
	cost := model.ChannelCost{}
	return cost.List(req.AppID, req.StartDate, req.EndDate)
}
