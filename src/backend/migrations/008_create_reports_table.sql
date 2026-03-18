-- 报表表创建脚本
-- 用于存储桥梁检测报表信息
-- 创建时间：2026-03-17

CREATE TABLE IF NOT EXISTS `reports` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '报表ID（主键）',
    `report_name` VARCHAR(200) NOT NULL COMMENT '报表名称',
    `report_type` ENUM('bridge_inspection', 'defect_analysis', 'health_comparison') NOT NULL COMMENT '报表类型',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '创建用户ID（外键）',
    `bridge_id` BIGINT UNSIGNED NULL COMMENT '关联桥梁ID（单桥梁报表）',
    `bridge_ids` JSON NULL COMMENT '关联桥梁ID列表（多桥梁报表）',
    `start_time` DATETIME NOT NULL COMMENT '报表开始时间',
    `end_time` DATETIME NOT NULL COMMENT '报表结束时间',
    `file_path` VARCHAR(500) NULL COMMENT 'PDF文件路径',
    `file_size` BIGINT NULL DEFAULT 0 COMMENT '文件大小（字节）',
    `status` ENUM('generating', 'completed', 'failed') NOT NULL DEFAULT 'generating' COMMENT '生成状态',
    `error_message` TEXT NULL COMMENT '错误信息（失败时）',
    `total_pages` INT NULL DEFAULT 0 COMMENT '总页数',
    `defect_count` INT NULL DEFAULT 0 COMMENT '缺陷数量',
    `high_risk_count` INT NULL DEFAULT 0 COMMENT '高危缺陷数量',
    `health_score` DECIMAL(5,2) NULL DEFAULT 0.00 COMMENT '健康度评分',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间（软删除）',

    -- 索引
    INDEX `idx_user_id` (`user_id`) COMMENT '用户ID索引',
    INDEX `idx_bridge_id` (`bridge_id`) COMMENT '桥梁ID索引',
    INDEX `idx_report_type` (`report_type`) COMMENT '报表类型索引',
    INDEX `idx_status` (`status`) COMMENT '状态索引',
    INDEX `idx_created_at` (`created_at`) COMMENT '创建时间索引',
    INDEX `idx_deleted_at` (`deleted_at`) COMMENT '软删除索引',

    -- 外键约束
    CONSTRAINT `fk_reports_user_id` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_reports_bridge_id` FOREIGN KEY (`bridge_id`) REFERENCES `bridges`(`id`) ON DELETE SET NULL ON UPDATE CASCADE

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='检测报表';
