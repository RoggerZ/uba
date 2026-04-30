package main

import (
	"errors"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestConsumeAsyncProducerErrorsReturnsAfterChannelClose(t *testing.T) {
	logs.Logger = zap.NewNop()

	errCh := make(chan *sarama.ProducerError, 1)
	done := make(chan struct{})

	go func() {
		consumeAsyncProducerErrors(errCh)
		close(done)
	}()

	errCh <- &sarama.ProducerError{
		Msg: &sarama.ProducerMessage{Topic: "report"},
		Err: errors.New("produce failed"),
	}
	close(errCh)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("consumer did not exit after error channel was closed")
	}
}
