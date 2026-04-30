package sinker

import (
	"bytes"

	"github.com/1340691923/xwl_bi/model"
)

func GetClusterSql() string {
	// GetClusterSql 统一拼接 on cluster 语句。
	//
	// 这样调用方不需要在每个建表/补列位置重复判断：
	// 1. 是否开启集群
	// 2. 开启后 clusterName 怎么拼
	if model.GlobConfig.Comm.ClickHouse.ClusterName == "" {
		return " "
	}

	b := bytes.Buffer{}
	b.WriteString(" on cluster ")
	b.WriteString(model.GlobConfig.Comm.ClickHouse.ClusterName)
	b.WriteString(" ")
	return b.String()
}

func GetMergeTree(tableName string) string {
	// GetMergeTree 根据当前部署模式返回单机或集群版 MergeTree 引擎定义。
	if model.GlobConfig.Comm.ClickHouse.ClusterName == "" {
		return "MergeTree"
	}
	return `ReplicatedMergeTree('/clickhouse/` + model.GlobConfig.Comm.ClickHouse.DbName + `/tables/{` + model.GlobConfig.Comm.ClickHouse.MacrosShardKeyName + `}/` + tableName + `', '{` + model.GlobConfig.Comm.ClickHouse.MacrosReplicaKeyName + `}')`
}

func GetReplacingMergeTree(tableName, ext string) string {
	// GetReplacingMergeTree 与 GetMergeTree 类似，只是用于需要版本列或替换语义的表。
	if model.GlobConfig.Comm.ClickHouse.ClusterName == "" {
		return "ReplacingMergeTree"
	}
	return `ReplicatedReplacingMergeTree('/clickhouse/` + model.GlobConfig.Comm.ClickHouse.DbName + `/tables/{` + model.GlobConfig.Comm.ClickHouse.MacrosShardKeyName + `}/` + tableName + `', '{` + model.GlobConfig.Comm.ClickHouse.MacrosReplicaKeyName + `}',` + ext + `)`
}
