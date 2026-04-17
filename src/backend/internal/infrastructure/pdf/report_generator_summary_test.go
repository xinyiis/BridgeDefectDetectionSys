package pdf

import (
	"testing"
	"time"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

func TestClassifyRiskAndNextInspection(t *testing.T) {
	risk, days, _ := classifyRiskAndNextInspection(20, 6, 30)
	if risk != "高风险" || days != 7 {
		t.Fatalf("expected high risk with 7 days, got risk=%s days=%d", risk, days)
	}

	risk, days, _ = classifyRiskAndNextInspection(65, 1, 8)
	if risk != "中风险" || days != 14 {
		t.Fatalf("expected medium risk with 14 days, got risk=%s days=%d", risk, days)
	}

	risk, days, _ = classifyRiskAndNextInspection(88, 0, 5)
	if risk != "低风险" || days != 30 {
		t.Fatalf("expected low risk with 30 days, got risk=%s days=%d", risk, days)
	}
}

func TestBuildTrendSummary(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 4, 11, 23, 59, 59, 0, time.Local)
	report := &model.Report{StartTime: start, EndTime: end}

	// 后半程明显增多，应该判断上升。
	defects := []model.Defect{
		{DetectedAt: start.Add(24 * time.Hour)},
		{DetectedAt: start.Add(48 * time.Hour)},
		{DetectedAt: start.Add(8 * 24 * time.Hour)},
		{DetectedAt: start.Add(9 * 24 * time.Hour)},
		{DetectedAt: start.Add(10 * 24 * time.Hour)},
		{DetectedAt: start.Add(10*24*time.Hour + 2*time.Hour)},
	}

	trend := buildTrendSummary(report, defects)
	if trend == "" {
		t.Fatal("expected non-empty trend summary")
	}
	if trend == "前后半程缺陷数接近（2 vs 4），整体平稳" {
		t.Fatalf("expected rising trend classification, got %q", trend)
	}
}

func TestBuildReportSummary_Basic(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 4, 17, 23, 59, 59, 0, time.Local)
	report := &model.Report{
		DefectCount:   10,
		HighRiskCount: 2,
		HealthScore:   68.5,
		StartTime:     start,
		EndTime:       end,
	}

	defects := []model.Defect{
		{DefectType: "裂缝", Confidence: 0.92, Area: 0.03, DetectedAt: start.Add(24 * time.Hour)},
		{DefectType: "裂缝", Confidence: 0.90, Area: 0.02, DetectedAt: start.Add(48 * time.Hour)},
		{DefectType: "剥落", Confidence: 0.85, Area: 0.01, DetectedAt: end.Add(-24 * time.Hour)},
	}

	summary := buildReportSummary(report, defects)
	if summary.HealthLevel != "一般" {
		t.Fatalf("expected health level 一般, got %s", summary.HealthLevel)
	}
	if summary.RiskLevel == "" {
		t.Fatal("expected non-empty risk level")
	}
	if len(summary.KeyFindings) < 3 {
		t.Fatalf("expected at least 3 key findings, got %d", len(summary.KeyFindings))
	}
	if len(summary.PriorityActions) < 3 {
		t.Fatalf("expected at least 3 priority actions, got %d", len(summary.PriorityActions))
	}
}
