package sinker

import (
	"context"
	"testing"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type fakeConsumerGroupSession struct {
	claims       map[string][]int32
	generationID int32
}

func (f *fakeConsumerGroupSession) Claims() map[string][]int32 {
	return f.claims
}

func (f *fakeConsumerGroupSession) MemberID() string {
	return "member"
}

func (f *fakeConsumerGroupSession) GenerationID() int32 {
	return f.generationID
}

func (f *fakeConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
}

func (f *fakeConsumerGroupSession) Commit() {}

func (f *fakeConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
}

func (f *fakeConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {}

func (f *fakeConsumerGroupSession) Context() context.Context {
	return context.Background()
}

func TestGetSaramaConfigEnablesConsumerErrors(t *testing.T) {
	cfg, err := GetSaramaConfig(model.KafkaCfg{})
	if err != nil {
		t.Fatalf("GetSaramaConfig returned error: %v", err)
	}
	if !cfg.Consumer.Return.Errors {
		t.Fatal("Consumer.Return.Errors = false, want true")
	}
}

func TestKafkaSaramaCleanupClearsMarkedOffsetsForClaims(t *testing.T) {
	oldLogger := logs.Logger
	logs.Logger = zap.NewNop()
	defer func() {
		logs.Logger = oldLogger
	}()

	kafka := NewKafkaSarama()
	kafka.topic = "test005"
	kafka.cleanupFn = func(generationID int32) {}
	kafka.storeMarkedOffset(0, 100)
	kafka.storeMarkedOffset(1, 200)
	kafka.storeMarkedOffset(2, 300)

	session := &fakeConsumerGroupSession{
		claims: map[string][]int32{
			"test005": {0, 2},
		},
		generationID: 1,
	}

	handler := MyConsumerGroupHandler{k: kafka}
	if err := handler.Cleanup(session); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}

	if got := kafka.CurrentMarkedOffset(); got != 200 {
		t.Fatalf("CurrentMarkedOffset after cleanup = %d, want 200", got)
	}
}
