package pannel

import (
	"encoding/json"
	"strconv"
	"strings"
)

const DefaultReportTableCardSize = "medium"

type ReportTableCard struct {
	Id   string `json:"id"`
	Size string `json:"size"`
}

type reportTableCardAlias struct {
	Id   interface{} `json:"id"`
	RtId interface{} `json:"rt_id"`
	Size string      `json:"size"`
}

func normalizeReportTableCardSize(size string) string {
	switch size {
	case "small", "medium", "large":
		return size
	default:
		return DefaultReportTableCardSize
	}
}

func normalizeReportTableCardID(value interface{}) string {
	switch id := value.(type) {
	case string:
		return strings.TrimSpace(id)
	case float64:
		return strconv.Itoa(int(id))
	case int:
		return strconv.Itoa(id)
	case int64:
		return strconv.Itoa(int(id))
	default:
		return ""
	}
}

func NormalizeReportTableCards(cards []ReportTableCard) []ReportTableCard {
	res := make([]ReportTableCard, 0, len(cards))
	seen := map[string]struct{}{}

	for _, card := range cards {
		card.Id = strings.TrimSpace(card.Id)
		if card.Id == "" {
			continue
		}

		if _, ok := seen[card.Id]; ok {
			continue
		}

		card.Size = normalizeReportTableCardSize(card.Size)
		res = append(res, card)
		seen[card.Id] = struct{}{}
	}

	return res
}

func ParseReportTableCards(raw string) []ReportTableCard {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []ReportTableCard{}
	}

	if strings.HasPrefix(raw, "[") {
		var cards []ReportTableCard
		if err := json.Unmarshal([]byte(raw), &cards); err == nil {
			return NormalizeReportTableCards(cards)
		}

		var aliases []reportTableCardAlias
		if err := json.Unmarshal([]byte(raw), &aliases); err == nil {
			cards = make([]ReportTableCard, 0, len(aliases))
			for _, alias := range aliases {
				id := normalizeReportTableCardID(alias.Id)
				if id == "" {
					id = normalizeReportTableCardID(alias.RtId)
				}
				cards = append(cards, ReportTableCard{
					Id:   id,
					Size: alias.Size,
				})
			}
			return NormalizeReportTableCards(cards)
		}
	}

	parts := strings.Split(raw, ",")
	cards := make([]ReportTableCard, 0, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		cards = append(cards, ReportTableCard{
			Id:   id,
			Size: DefaultReportTableCardSize,
		})
	}

	return NormalizeReportTableCards(cards)
}

func SerializeReportTableCards(cards []ReportTableCard) string {
	cards = NormalizeReportTableCards(cards)
	if len(cards) == 0 {
		return ""
	}

	data, err := json.Marshal(cards)
	if err != nil {
		return ""
	}
	return string(data)
}

func AppendReportTableCard(cards []ReportTableCard, targetId int, size string) []ReportTableCard {
	target := strconv.Itoa(targetId)
	res := NormalizeReportTableCards(cards)

	for index, card := range res {
		if card.Id != target {
			continue
		}

		res[index].Size = normalizeReportTableCardSize(size)
		return NormalizeReportTableCards(res)
	}

	res = append(res, ReportTableCard{
		Id:   target,
		Size: normalizeReportTableCardSize(size),
	})

	return NormalizeReportTableCards(res)
}

func RemoveReportTableCard(cards []ReportTableCard, targetId int) []ReportTableCard {
	target := strconv.Itoa(targetId)
	res := make([]ReportTableCard, 0, len(cards))

	for _, card := range cards {
		if card.Id == target {
			continue
		}
		res = append(res, card)
	}

	return NormalizeReportTableCards(res)
}
