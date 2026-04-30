package report

// ReportService 保留为兼容入口，内部转发到拆分后的 resolver、membership、producer。
//
// 这样做的原因是：
// 1. 旧调用方如果还在使用零值 ReportService{}，不会被立刻打断。
// 2. 新代码则应该优先直接依赖更小职责的对象，而不是继续把所有能力堆在一个 service 上。
type ReportService struct {
	Resolver        *TableIDResolver
	DebugMembership *DebugMembershipChecker
	DebugProducer   *DebugDataProducer
}

func (s *ReportService) resolver() *TableIDResolver {
	if s != nil && s.Resolver != nil {
		return s.Resolver
	}
	return DefaultTableIDResolver()
}

func (s *ReportService) debugMembership() *DebugMembershipChecker {
	if s != nil && s.DebugMembership != nil {
		return s.DebugMembership
	}
	return NewDebugMembershipChecker(nil)
}

func (s *ReportService) debugProducer() *DebugDataProducer {
	if s != nil && s.DebugProducer != nil {
		return s.DebugProducer
	}
	return NewDefaultDebugDataProducer()
}

// GetTableid 保留旧方法名，兼容历史调用。
func (s *ReportService) GetTableid(appid, appkey string) (string, error) {
	return s.resolver().Resolve(appid, appkey)
}

// IsDebugUser 保留旧方法名，内部统一转发到 IsDebugDevice。
func (s *ReportService) IsDebugUser(debug, xwlDistinctID, tableID string) bool {
	return s.debugMembership().IsDebugDevice(debug, xwlDistinctID, tableID)
}

// CantInflowOfKakfa 保留旧方法名，兼容老代码中的拼写。
func (s *ReportService) CantInflowOfKakfa(debug, xwlDistinctID, tableID string) bool {
	return s.debugMembership().IsDebugDevice(debug, xwlDistinctID, tableID)
}

// InflowOfDebugData 保留旧入口，内部转发到新的 debug producer。
func (s *ReportService) InflowOfDebugData(data map[string]interface{}, eventName string) error {
	return s.debugProducer().Send(data)
}
