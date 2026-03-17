-- ========================================
-- 桥梁缺陷检测系统 - 统计模块索引优化
-- ========================================

USE bridge_detection;

-- 1. 检查现有索引
SHOW INDEX FROM defects;

-- 2. 创建缺陷表索引（如果不存在）

-- 桥梁ID索引（权限过滤）
CREATE INDEX IF NOT EXISTS idx_defects_bridge_id ON defects(bridge_id);

-- 检测时间索引（时间范围查询）
CREATE INDEX IF NOT EXISTS idx_defects_detected_at ON defects(detected_at);

-- 缺陷类型索引（类型分布统计）
CREATE INDEX IF NOT EXISTS idx_defects_type ON defects(defect_type);

-- 软删除索引（过滤已删除数据）
CREATE INDEX IF NOT EXISTS idx_defects_deleted_at ON defects(deleted_at);

-- 组合索引（权限 + 时间查询）
CREATE INDEX IF NOT EXISTS idx_defects_bridge_time ON defects(bridge_id, detected_at);

-- 置信度索引（高危告警查询）
CREATE INDEX IF NOT EXISTS idx_defects_confidence ON defects(confidence);

-- 3. 验证索引创建
SHOW INDEX FROM defects;

-- 4. 分析索引效果
EXPLAIN SELECT COUNT(*) FROM defects
WHERE bridge_id = 1
  AND detected_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
  AND deleted_at IS NULL;
