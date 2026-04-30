package model

import "time"

type InputMessage struct {
	Topic     string
	Partition int
	Key       []byte
	Value     []byte
	Offset    int64
	Timestamp *time.Time
	// GenerationID 记录这条消息所属的 consumer session，
	// 便于把 mark 行为和 session cleanup 日志串起来。
	GenerationID int32
}
