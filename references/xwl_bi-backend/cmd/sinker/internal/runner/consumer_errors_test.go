package runner

import (
	"errors"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"go.uber.org/zap"
)

func TestConsumeKafkaConsumerErrorsReturnsAfterChannelClose(t *testing.T) {
	logs.Logger = zap.NewNop()

	errCh := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		consumeKafkaConsumerErrors("report-consumer", "report-topic", errCh)
		close(done)
	}()

	errCh <- errors.New("consume failed")
	close(errCh)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("consumer error loop did not exit after channel close")
	}
}
