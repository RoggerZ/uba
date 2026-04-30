package report

import (
	"sync"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/my_error"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
)

// PayloadBuildInput 表示一次上报消息构造所需的全部输入。
//
// 设计原因：
// 1. 把构造 KafkaData 所需字段集中成一个输入对象，避免新增字段时到处改函数签名。
// 2. user/event 两类上报只有少数字段差异，统一输入后更容易通过注册扩展。
//
// 示例：
// 1. typ=reportUser 时，builder 会把 EventName 固定为“用户属性”
// 2. typ=reportEvent 时，builder 会保留业务传入的 eventName
type PayloadBuildInput struct {
	APPID              string
	TableID            string
	Debug              string
	ReportTime         string
	ReportTimeHasClock bool
	EventName          string
	IP                 string
	Body               []byte
}

// PayloadBuilder 负责把上报输入转换为 KafkaData。
type PayloadBuilder func(input PayloadBuildInput) model.KafkaData

// PayloadBuilderRegistry 负责维护“上报类型 -> 构造器”的注册关系。
//
// 这里使用注册表而不是 if/else 的原因是：
// 1. 新增上报类型时，只需要注册新的 builder。
// 2. 稳定主流程不必每次跟着修改，符合开闭原则。
// 3. builder 只负责把输入变成 KafkaData，不负责发送，这样领域对象和基础设施职责更清晰。
type PayloadBuilderRegistry struct {
	mu       sync.RWMutex
	builders map[string]PayloadBuilder
}

func NewPayloadBuilderRegistry() *PayloadBuilderRegistry {
	return &PayloadBuilderRegistry{
		builders: make(map[string]PayloadBuilder),
	}
}

// Register 注册某种上报类型对应的 payload builder。
func (r *PayloadBuilderRegistry) Register(typ string, builder PayloadBuilder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builders[typ] = builder
}

// Build 根据 typ 构造对应的 KafkaData。
//
// 示例：
// 1. typ=reportUser -> ReportType=UserReportType, EventName=用户属性
// 2. typ=reportEvent -> ReportType=EventReportType, EventName 保持业务传入值
// 3. typ 非法 -> 返回现有业务错误“上报类型错误”
func (r *PayloadBuilderRegistry) Build(typ string, input PayloadBuildInput) (model.KafkaData, error) {
	r.mu.RLock()
	builder, ok := r.builders[typ]
	r.mu.RUnlock()
	if !ok {
		return model.KafkaData{}, my_error.NewBusiness(ERROR_TABLE, ReportTypeErr)
	}
	return builder(input), nil
}

var defaultPayloadBuilderRegistry = newDefaultPayloadBuilderRegistry()

func newDefaultPayloadBuilderRegistry() *PayloadBuilderRegistry {
	registry := NewPayloadBuilderRegistry()
	registry.Register(model2.ReportUserProperties, buildUserPayload)
	registry.Register(model2.ReportEventProperties, buildEventPayload)
	return registry
}

// DefaultPayloadBuilderRegistry 返回 report 链路默认使用的 payload 注册表。
func DefaultPayloadBuilderRegistry() *PayloadBuilderRegistry {
	return defaultPayloadBuilderRegistry
}

func buildUserPayload(input PayloadBuildInput) model.KafkaData {
	reportTimeHasClock := input.ReportTimeHasClock
	return model.KafkaData{
		APPID:              input.APPID,
		TableId:            input.TableID,
		Ip:                 input.IP,
		Debug:              input.Debug,
		ReqData:            input.Body,
		ReportTime:         input.ReportTime,
		ReportTimeHasClock: &reportTimeHasClock,
		ReportType:         model.UserReportType,
		EventName:          "用户属性",
	}
}

func buildEventPayload(input PayloadBuildInput) model.KafkaData {
	reportTimeHasClock := input.ReportTimeHasClock
	return model.KafkaData{
		APPID:              input.APPID,
		TableId:            input.TableID,
		Ip:                 input.IP,
		Debug:              input.Debug,
		ReqData:            input.Body,
		ReportTime:         input.ReportTime,
		ReportTimeHasClock: &reportTimeHasClock,
		ReportType:         model.EventReportType,
		EventName:          input.EventName,
	}
}
