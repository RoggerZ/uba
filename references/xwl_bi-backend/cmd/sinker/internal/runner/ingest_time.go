package runner

import (
	"time"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func resolveIngestTime(decoded DecodedMessage) string {
	return time.Now().Local().Format(util.TimeFormat)
}
