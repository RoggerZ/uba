package runner

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
)

type fakeConsumerOffsetReader struct {
	offsets map[int32]int64
}

func (f *fakeConsumerOffsetReader) CurrentMarkedOffset() int64 {
	var total int64
	for _, offset := range f.offsets {
		total += offset
	}
	return total
}

func (f *fakeConsumerOffsetReader) CurrentMarkedOffsets() map[int32]int64 {
	result := make(map[int32]int64, len(f.offsets))
	for partition, offset := range f.offsets {
		result[partition] = offset
	}
	return result
}

type fakeConsumerGroupOffsetAdmin struct {
	response *sarama.OffsetFetchResponse
	err      error
}

func (f *fakeConsumerGroupOffsetAdmin) ListConsumerGroupOffsets(group string, topicPartitions map[string][]int32) (*sarama.OffsetFetchResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

func (f *fakeConsumerGroupOffsetAdmin) Close() error {
	return nil
}

func TestConsumerRateSamplerCurrentOffsetUsesCommittedOffsetsOnColdStart(t *testing.T) {
	sampler := consumerRateSampler{
		group:    "group",
		topic:    "topic",
		consumer: &fakeConsumerOffsetReader{offsets: map[int32]int64{}},
		admin: &fakeConsumerGroupOffsetAdmin{
			response: buildOffsetFetchResponse("topic", map[int32]int64{
				0: 101,
				1: 205,
			}),
		},
	}

	got := sampler.currentOffset([]int32{0, 1})
	if got != 306 {
		t.Fatalf("currentOffset() = %d, want 306", got)
	}
}

func TestConsumerRateSamplerCurrentOffsetPrefersMarkedOffsetsWhenAhead(t *testing.T) {
	sampler := consumerRateSampler{
		group: "group",
		topic: "topic",
		consumer: &fakeConsumerOffsetReader{offsets: map[int32]int64{
			0: 120,
			1: 210,
		}},
		admin: &fakeConsumerGroupOffsetAdmin{
			response: buildOffsetFetchResponse("topic", map[int32]int64{
				0: 101,
				1: 205,
			}),
		},
	}

	got := sampler.currentOffset([]int32{0, 1})
	if got != 330 {
		t.Fatalf("currentOffset() = %d, want 330", got)
	}
}

func TestConsumerRateSamplerCurrentOffsetFallsBackToMarkedOffsetsWhenAdminFails(t *testing.T) {
	sampler := consumerRateSampler{
		group: "group",
		topic: "topic",
		consumer: &fakeConsumerOffsetReader{offsets: map[int32]int64{
			0: 120,
			1: 210,
		}},
		admin: &fakeConsumerGroupOffsetAdmin{
			err: errors.New("offset fetch failed"),
		},
	}

	got := sampler.currentOffset([]int32{0, 1})
	if got != 330 {
		t.Fatalf("currentOffset() = %d, want 330", got)
	}
}

func buildOffsetFetchResponse(topic string, offsets map[int32]int64) *sarama.OffsetFetchResponse {
	response := &sarama.OffsetFetchResponse{
		Blocks: map[string]map[int32]*sarama.OffsetFetchResponseBlock{
			topic: {},
		},
	}
	for partition, offset := range offsets {
		response.Blocks[topic][partition] = &sarama.OffsetFetchResponseBlock{
			Offset: offset,
			Err:    sarama.ErrNoError,
		}
	}
	return response
}
