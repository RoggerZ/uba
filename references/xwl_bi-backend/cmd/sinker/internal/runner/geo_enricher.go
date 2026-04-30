package runner

import (
	"github.com/1340691923/xwl_bi/cmd/sinker/geoip"
	"github.com/tidwall/sjson"
)

type geoEnricher struct {
	geo *geoip.Geoip2
}

// newGeoEnricher 创建一个“IP -> 地理字段补充器”。
func newGeoEnricher(geoResolver *geoip.Geoip2) *geoEnricher {
	return &geoEnricher{geo: geoResolver}
}

// Enrich 根据 IP 给 ReqData 增补地理字段。
//
// 处理原则：
// 1. 先补 xwl_ip，保证最基础的服务端信息一定存在。
// 2. 再尽力补国家、省份、城市、运营商、ASN。
// 3. 即使 GeoIP 查询失败，也不把原始消息判失败，只返回当前可保留的数据。
//
// 示例：
// 1. 输入 reqData={"a":1}, ip="8.8.8.8"
// 2. 输出至少会带上 xwl_ip，若 GeoIP 查询成功，还会继续带上 xwl_country/xwl_city 等字段
func (e *geoEnricher) Enrich(reqData []byte, ip string) ([]byte, error) {
	if ip == "" {
		return reqData, nil
	}

	// 即使地理信息查询失败，也保留原始 ip 字段，避免完全丢掉最基础的服务端补充信息。
	enriched, _ := sjson.SetBytes(reqData, "xwl_ip", ip)
	geoInfo, err := e.geo.GetGeoInfoFromIP(ip)
	if err != nil {
		return enriched, err
	}

	if geoInfo.Province != "" {
		enriched, _ = sjson.SetBytes(enriched, "xwl_province", geoInfo.Province)
	}
	if geoInfo.City != "" {
		enriched, _ = sjson.SetBytes(enriched, "xwl_city", geoInfo.City)
	}
	if geoInfo.Country != "" {
		enriched, _ = sjson.SetBytes(enriched, "xwl_country", geoInfo.Country)
	}
	if geoInfo.ISP != "" {
		enriched, _ = sjson.SetBytes(enriched, "xwl_isp", geoInfo.ISP)
	}
	if geoInfo.ASN != "" {
		enriched, _ = sjson.SetBytes(enriched, "xwl_asn", geoInfo.ASN)
	}

	return enriched, nil
}
