-- 用户分群表现网迁移脚本
-- 适用场景：
-- 1. 现网已存在 user_group 表
-- 2. 需要补齐规则型/静态快照型用户分群相关字段

ALTER TABLE `user_group`
  ADD COLUMN `group_display_name` varchar(255) COLLATE utf8mb4_german2_ci NOT NULL DEFAULT '' AFTER `group_name`,
  ADD COLUMN `update_type` varchar(32) COLLATE utf8mb4_german2_ci NOT NULL DEFAULT 'manual' AFTER `group_remark`,
  ADD COLUMN `create_type` varchar(64) COLLATE utf8mb4_german2_ci NOT NULL DEFAULT 'analysis_page_snapshot' AFTER `update_type`,
  ADD COLUMN `rule_content` longtext COLLATE utf8mb4_german2_ci NULL AFTER `create_type`,
  ADD COLUMN `last_calculate_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `user_list`,
  ADD COLUMN `can_manual_refresh` tinyint(1) NOT NULL DEFAULT '0' AFTER `last_calculate_time`;

-- 历史数据回填
UPDATE `user_group`
SET
  `group_display_name` = `group_name`
WHERE
  (`group_display_name` IS NULL OR `group_display_name` = '');

UPDATE `user_group`
SET
  `last_calculate_time` = `update_time`
WHERE
  `last_calculate_time` IS NULL;

-- 新增唯一索引
ALTER TABLE `user_group`
  ADD UNIQUE KEY `user_group_display_name` (`group_display_name`,`appid`) USING BTREE;
