package report

import (
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
	"github.com/IBM/sarama"
	jsoniter "github.com/json-iterator/go"
)

func TestJSONTopicProducerSend(t *testing.T) {
	var captured *sarama.ProducerMessage

	producer := newJSONTopicProducerWithSender(func(msg *sarama.ProducerMessage) error {
		captured = msg
		return nil
	})
	producer.now = func() time.Time {
		return time.Date(2026, 4, 10, 12, 0, 0, 0, time.Local)
	}

	payload := map[string]interface{}{
		"code": 1,
		"msg":  "ok",
	}
	if err := producer.Send("report-topic", payload); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if captured == nil {
		t.Fatal("expected producer message to be captured")
	}
	if captured.Topic != "report-topic" {
		t.Fatalf("topic = %q, want %q", captured.Topic, "report-topic")
	}
	if !captured.Timestamp.Equal(producer.now()) {
		t.Fatalf("timestamp = %v, want %v", captured.Timestamp, producer.now())
	}

	bytes, err := captured.Value.Encode()
	if err != nil {
		t.Fatalf("Value.Encode returned error: %v", err)
	}

	var decoded map[string]interface{}
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if decoded["msg"] != "ok" {
		t.Fatalf("msg = %v, want %q", decoded["msg"], "ok")
	}
}

func TestResolveProducerType(t *testing.T) {
	oldType := model.GlobConfig.Comm.Kafka.ProducerType
	defer func() {
		model.GlobConfig.Comm.Kafka.ProducerType = oldType
	}()

	model.GlobConfig.Comm.Kafka.ProducerType = ""
	if got := resolveProducerType(); got != "sync" {
		t.Fatalf("resolveProducerType() with empty config = %q, want %q", got, "sync")
	}

	model.GlobConfig.Comm.Kafka.ProducerType = "async"
	if got := resolveProducerType(); got != "async" {
		t.Fatalf("resolveProducerType() with async config = %q, want %q", got, "async")
	}
}
