package ck

import (
	"fmt"
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker"
	"log"
	"strconv"
)

// Init 初始化clickhouse 表数据
func Init() {
	var err error

	_, err = db.ClickHouseSqlx.Exec(` create database if not exists ` + model.GlobConfig.Comm.ClickHouse.DbName + ` ` + sinker.GetClusterSql())
	if err != nil {
		log.Println(fmt.Sprintf("clickhouse 建库 "+model.GlobConfig.Comm.ClickHouse.DbName+" 失败:%s", err.Error()))
		panic(err)
	}

	_, err = db.ClickHouseSqlx.Exec(`DROP TABLE IF EXISTS ` + consumer_data.TableNameAcceptanceStatus + sinker.GetClusterSql() + `;`)
	if err != nil {
		log.Println(fmt.Sprintf("clickhouse 删除表 %s 失败:%s", consumer_data.TableNameAcceptanceStatus, err.Error()))
		panic(err)
	}

	_, err = db.ClickHouseSqlx.Exec(`
		CREATE TABLE ` + consumer_data.TableNameAcceptanceStatus + ` ` + sinker.GetClusterSql() + `
		(
			table_id Int64,
			ingest_time DateTime,
			part_date DateTime DEFAULT now(),
			data_name String,
			error_reason String,
			error_handling String,
			report_type String,
			report_data String,
			xwl_kafka_offset Int64,
			part_event String,
			status Int32
		)
		ENGINE = ` + sinker.GetMergeTree(consumer_data.TableNameAcceptanceStatus) + `
		PARTITION BY (toYYYYMM(ingest_time))
		ORDER BY (
		 table_id,
		 part_date,
		 data_name,
		 error_reason,
		 error_handling,
		 report_type,
		 status)
		TTL ingest_time + toIntervalMonth(` + strconv.Itoa(consumer_data.SidecarRetentionMonths()) + `)
		SETTINGS index_granularity = 8192;
`)
	if err != nil {
		log.Println(fmt.Sprintf("clickhouse 建表 %s 失败:%s", consumer_data.TableNameAcceptanceStatus, err.Error()))
		panic(err)
	}

	_, err = db.ClickHouseSqlx.Exec(`DROP TABLE IF EXISTS ` + consumer_data.TableNameRealTimeWarehousing + sinker.GetClusterSql() + `;`)

	if err != nil {
		log.Println(fmt.Sprintf("clickhouse 删除表 %s 失败:%s", consumer_data.TableNameRealTimeWarehousing, err.Error()))
		panic(err)
	}

	_, err = db.ClickHouseSqlx.Exec(`
		CREATE TABLE ` + consumer_data.TableNameRealTimeWarehousing + ` ` + sinker.GetClusterSql() + `
		(
		
			table_id Int64,
			ingest_time DateTime,
			event_time DateTime,
			event_name String,
			report_data String
		)
		ENGINE = ` + sinker.GetMergeTree(consumer_data.TableNameRealTimeWarehousing) + ` 
		PARTITION BY (toYYYYMM(ingest_time))
		ORDER BY (
		 table_id,
		 event_time,
		 event_name)
		TTL ingest_time + toIntervalMonth(` + strconv.Itoa(consumer_data.SidecarRetentionMonths()) + `)
		SETTINGS index_granularity = 8192;
`)
	if err != nil {
		log.Println(fmt.Sprintf("clickhouse 建表 %s 失败:%s", consumer_data.TableNameRealTimeWarehousing, err.Error()))
		panic(err)
	}

	log.Println("初始化CK数据完成！")
}
