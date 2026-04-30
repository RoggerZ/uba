package action

import (
	"sync"
	"testing"

	"github.com/1340691923/xwl_bi/model"
)

func TestAddMetaEventOnlyQueuesOncePerEvent(t *testing.T) {
	resetMetaEventState()

	kafkaData := model.KafkaData{
		TableId:    "51",
		EventName:  "AppLaunch",
		ReportType: model.EventReportType,
	}

	if err := AddMetaEvent(kafkaData); err != nil {
		t.Fatalf("first AddMetaEvent returned error: %v", err)
	}
	if err := AddMetaEvent(kafkaData); err != nil {
		t.Fatalf("second AddMetaEvent returned error: %v", err)
	}

	if got := len(metaEventChan); got != 1 {
		t.Fatalf("queued meta events = %d, want 1", got)
	}
}

func TestAddMetaEventIgnoresNonEventReportType(t *testing.T) {
	resetMetaEventState()

	if err := AddMetaEvent(model.KafkaData{
		TableId:    "51",
		EventName:  "UserProfile",
		ReportType: model.UserReportType,
	}); err != nil {
		t.Fatalf("AddMetaEvent returned error: %v", err)
	}

	if got := len(metaEventChan); got != 0 {
		t.Fatalf("queued meta events = %d, want 0", got)
	}
}

func resetMetaEventState() {
	MetaEventMap = sync.Map{}
	for len(metaEventChan) > 0 {
		<-metaEventChan
	}
}
