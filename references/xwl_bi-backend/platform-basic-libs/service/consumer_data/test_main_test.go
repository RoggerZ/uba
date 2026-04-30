package consumer_data

import (
	"os"
	"testing"

	"github.com/1340691923/xwl_bi/engine/logs"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logs.Logger = zap.NewNop()
	os.Exit(m.Run())
}
