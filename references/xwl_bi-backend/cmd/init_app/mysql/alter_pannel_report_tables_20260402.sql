-- Change pannel.report_tables to text to avoid varchar(255) overflow when saving dashboard card config

ALTER TABLE `pannel`
  MODIFY COLUMN `report_tables` text CHARACTER SET utf8mb4 COLLATE utf8mb4_german2_ci NULL;

UPDATE `pannel`
SET `report_tables` = ''
WHERE `report_tables` IS NULL;
