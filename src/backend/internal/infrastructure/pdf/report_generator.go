// Package pdf 实现PDF报表生成功能
package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/model"
)

// ReportGenerator PDF报表生成器
type ReportGenerator struct {
	pdf      *gopdf.GoPdf
	fontPath string
}

// NewReportGenerator 创建报表生成器实例
func NewReportGenerator(fontPath string) *ReportGenerator {
	return &ReportGenerator{
		fontPath: fontPath,
	}
}

// GenerateBridgeInspectionReport 生成桥梁检测报表
func (g *ReportGenerator) GenerateBridgeInspectionReport(
	report *model.Report,
	bridge *model.Bridge,
	defects []model.Defect,
	outputPath string,
) error {
	// 初始化PDF
	g.pdf = &gopdf.GoPdf{}
	g.pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	g.pdf.SetMargins(20, 20, 20, 20)

	// 添加中文字体
	if err := g.addChineseFont(); err != nil {
		return fmt.Errorf("添加中文字体失败: %v", err)
	}

	// 生成各个章节
	if err := g.generateCoverPage(report, bridge); err != nil {
		return fmt.Errorf("生成封面失败: %v", err)
	}
	if err := g.generateBridgeInfo(bridge); err != nil {
		return fmt.Errorf("生成桥梁信息失败: %v", err)
	}
	if err := g.generateDetectionOverview(report, defects); err != nil {
		return fmt.Errorf("生成检测概览失败: %v", err)
	}

	// 仅在存在缺陷数据时生成详情章节，避免大量空白页。
	if len(defects) > 0 {
		if err := g.generateStatistics(report, defects); err != nil {
			return fmt.Errorf("生成统计分析失败: %v", err)
		}

		if countHighRiskDefects(defects) > 0 {
			if err := g.generateHighRiskDefects(defects); err != nil {
				return fmt.Errorf("生成高危缺陷列表失败: %v", err)
			}
		}

		if err := g.generateDefectDetails(defects); err != nil {
			return fmt.Errorf("生成缺陷详情失败: %v", err)
		}
	}
	if err := g.generateConclusion(report, defects); err != nil {
		return fmt.Errorf("生成结论失败: %v", err)
	}

	// 保存PDF文件
	if err := g.pdf.WritePdf(outputPath); err != nil {
		return fmt.Errorf("保存PDF文件失败: %v", err)
	}

	return nil
}

// addChineseFont 添加中文字体
func (g *ReportGenerator) addChineseFont() error {
	candidates := buildFontCandidates(g.fontPath, os.Getenv("BRIDGE_PDF_FONT_PATH"))
	var loadErrors []string

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}

		ok, err := probeChineseFont(candidate)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("%s (探测失败: %v)", candidate, err))
			continue
		}
		if !ok {
			loadErrors = append(loadErrors, fmt.Sprintf("%s (不支持中文字符渲染)", candidate))
			continue
		}

		if err := g.registerFont("sourcehansans", candidate); err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("%s (%v)", candidate, err))
			continue
		}

		if err := g.pdf.SetFont("sourcehansans", "", 12); err != nil {
			return fmt.Errorf("设置字体失败: %v", err)
		}

		return nil
	}

	baseError := "未找到可用中文字体，请安装后重试"
	if len(loadErrors) > 0 {
		baseError = "中文字体存在但加载失败，请检查字体文件格式"
	}

	loadErrorText := "无（字体文件不存在）"
	if len(loadErrors) > 0 {
		loadErrorText = strings.Join(loadErrors, "; ")
	}

	return fmt.Errorf("%s。已尝试路径: %s。加载错误: %s。建议: 安装中文字体并配置 BRIDGE_PDF_FONT_PATH，例如 `sudo apt-get update && sudo apt-get install -y fonts-noto-cjk && export BRIDGE_PDF_FONT_PATH=/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc`", baseError, strings.Join(candidates, ", "), loadErrorText)
}

func (g *ReportGenerator) registerFont(fontName, fontPath string) error {
	addErr := g.pdf.AddTTFFont(fontName, fontPath)
	if addErr == nil {
		return nil
	}

	// gopdf 对 .ttc 支持有限，尝试通过字体数据方式兜底。
	if strings.EqualFold(filepath.Ext(fontPath), ".ttc") {
		data, readErr := os.ReadFile(fontPath)
		if readErr != nil {
			return fmt.Errorf("读取TTC字体文件失败: %v", readErr)
		}
		if len(data) < 12 {
			return fmt.Errorf("TTC文件格式无效")
		}
		if err := g.pdf.AddTTFFontData(fontName, data); err == nil {
			return nil
		}
	}

	return fmt.Errorf("加载字体文件失败: %s (%v)", fontPath, addErr)
}

func buildFontCandidates(configuredFontPath, envFontPath string) []string {
	candidates := make([]string, 0, 8)
	seen := make(map[string]struct{})

	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		candidates = append(candidates, path)
	}

	add(envFontPath)
	add(configuredFontPath)
	if configuredFontPath != "" {
		add(filepath.Join(filepath.Dir(configuredFontPath), "wqy-microhei.ttc"))
	}

	// 常见服务器内置中文字体路径
	add("/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc")
	add("/usr/share/fonts/opentype/noto/NotoSansCJK-Medium.ttc")
	add("/usr/share/fonts/truetype/wqy/wqy-microhei.ttc")

	return candidates
}

func probeChineseFont(fontPath string) (bool, error) {
	probe := &gopdf.GoPdf{}
	probe.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	probe.AddPage()

	if err := registerFontOnPDF(probe, "probe-font", fontPath); err != nil {
		return false, err
	}
	if err := probe.SetFont("probe-font", "", 12); err != nil {
		return false, fmt.Errorf("设置探测字体失败: %v", err)
	}

	width, err := probe.MeasureTextWidth("桥梁检测")
	if err != nil {
		return false, fmt.Errorf("探测字体测宽失败: %v", err)
	}

	// 宽度接近 0 说明中文字符未被字体正确支持。
	return width > 0.01, nil
}

func registerFontOnPDF(pdf *gopdf.GoPdf, fontName, fontPath string) error {
	addErr := pdf.AddTTFFont(fontName, fontPath)
	if addErr == nil {
		return nil
	}

	// gopdf 对 .ttc 支持有限，尝试通过字体数据方式兜底。
	if strings.EqualFold(filepath.Ext(fontPath), ".ttc") {
		data, readErr := os.ReadFile(fontPath)
		if readErr != nil {
			return fmt.Errorf("读取TTC字体文件失败: %v", readErr)
		}
		if len(data) < 12 {
			return fmt.Errorf("TTC文件格式无效")
		}
		if err := pdf.AddTTFFontData(fontName, data); err == nil {
			return nil
		}
	}

	return fmt.Errorf("加载字体文件失败: %s (%v)", fontPath, addErr)
}

// generateCoverPage 生成封面
func (g *ReportGenerator) generateCoverPage(report *model.Report, bridge *model.Bridge) error {
	g.pdf.AddPage()

	// 标题
	if err := g.pdf.SetFont("sourcehansans", "", 28); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	g.pdf.SetY(80)
	if err := g.pdf.Cell(nil, report.ReportName); err != nil {
		return err
	}

	// 报表类型
	g.pdf.SetY(110)
	if err := g.pdf.SetFont("sourcehansans", "", 16); err != nil {
		return err
	}
	g.pdf.SetTextColor(108, 117, 125)
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "桥梁检测报表"); err != nil {
		return err
	}

	// 分隔线
	g.pdf.SetStrokeColor(200, 200, 200)
	g.pdf.SetLineWidth(0.5)
	g.pdf.Line(50, 125, 160, 125)

	// 报表信息
	if err := g.pdf.SetFont("sourcehansans", "", 12); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)

	infoY := 145.0
	lineHeight := 10.0

	info := []struct {
		label string
		value string
	}{
		{"桥梁名称", bridge.BridgeName},
		{"桥梁编号", bridge.BridgeCode},
		{"检测时间", fmt.Sprintf("%s 至 %s",
			report.StartTime.Format("2006-01-02"),
			report.EndTime.Format("2006-01-02"))},
		{"生成时间", report.CreatedAt.Format("2006-01-02 15:04:05")},
		{"健康度评分", fmt.Sprintf("%.2f 分", report.HealthScore)},
	}

	for _, item := range info {
		g.pdf.SetX(50)
		g.pdf.SetY(infoY)
		g.pdf.SetTextColor(108, 117, 125)
		if err := g.pdf.Cell(nil, item.label+": "); err != nil {
			return err
		}
		g.pdf.SetX(90)
		g.pdf.SetTextColor(33, 37, 41)
		if err := g.pdf.Cell(nil, item.value); err != nil {
			return err
		}
		infoY += lineHeight
	}

	// 页脚（系统信息）
	g.pdf.SetY(260)
	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(173, 181, 189)
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "桥梁缺陷检测系统"); err != nil {
		return err
	}
	g.pdf.SetY(270)
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "Bridge Defect Detection System"); err != nil {
		return err
	}

	return nil
}

// generateBridgeInfo 生成桥梁信息章节
func (g *ReportGenerator) generateBridgeInfo(bridge *model.Bridge) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("1. 桥梁基本信息"); err != nil {
		return err
	}

	// 桥梁信息表格
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}

	infoData := []struct {
		label string
		value string
	}{
		{"桥梁名称", bridge.BridgeName},
		{"桥梁编号", bridge.BridgeCode},
		{"桥梁位置", bridge.Address},
		{"桥梁类型", bridge.BridgeType},
		{"建造年份", fmt.Sprintf("%d 年", bridge.BuildYear)},
		{"桥梁长度", fmt.Sprintf("%.2f 米", bridge.Length)},
		{"桥梁宽度", fmt.Sprintf("%.2f 米", bridge.Width)},
		{"经度", fmt.Sprintf("%.6f", bridge.Longitude)},
		{"纬度", fmt.Sprintf("%.6f", bridge.Latitude)},
		{"桥梁状态", bridge.Status},
	}

	// 绘制表格
	g.pdf.SetStrokeColor(222, 226, 230)
	g.pdf.SetLineWidth(0.1)
	currentY := 60.0

	for _, item := range infoData {
		g.pdf.SetY(currentY)
		// 标签列（浅灰色背景）
		g.pdf.SetFillColor(248, 249, 250)
		g.pdf.RectFromUpperLeftWithStyle(20, currentY, 60, 8, "F")
		g.pdf.SetX(22)
		g.pdf.SetY(currentY + 2)
		g.pdf.SetTextColor(33, 37, 41)
		if err := g.pdf.Cell(nil, item.label); err != nil {
			return err
		}

		// 值列（白色背景）
		g.pdf.SetFillColor(255, 255, 255)
		g.pdf.RectFromUpperLeftWithStyle(80, currentY, 110, 8, "F")
		g.pdf.SetX(82)
		g.pdf.SetY(currentY + 2)
		if err := g.pdf.Cell(nil, item.value); err != nil {
			return err
		}

		// 边框
		g.pdf.SetStrokeColor(222, 226, 230)
		g.pdf.RectFromUpperLeft(20, currentY, 60, 8)
		g.pdf.RectFromUpperLeft(80, currentY, 110, 8)

		currentY += 8
	}

	// 备注信息
	if bridge.Remark != "" {
		g.pdf.SetY(currentY + 5)
		if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
			return err
		}
		g.pdf.SetTextColor(108, 117, 125)
		g.pdf.SetX(20)
		if err := g.pdf.Cell(nil, "备注: "+bridge.Remark); err != nil {
			return err
		}
	}

	return nil
}

// generateDetectionOverview 生成检测概览章节
func (g *ReportGenerator) generateDetectionOverview(report *model.Report, defects []model.Defect) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("2. 检测概览"); err != nil {
		return err
	}

	// 统计数据
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)

	statsData := []struct {
		label string
		value string
		color []int // RGB颜色
	}{
		{"检测时间范围", fmt.Sprintf("%s 至 %s",
			report.StartTime.Format("2006-01-02"),
			report.EndTime.Format("2006-01-02")), []int{108, 117, 125}},
		{"缺陷总数", fmt.Sprintf("%d 个", report.DefectCount), []int{13, 110, 253}},
		{"高危缺陷数量", fmt.Sprintf("%d 个", report.HighRiskCount), []int{220, 53, 69}},
		{"健康度评分", fmt.Sprintf("%.2f 分", report.HealthScore), getHealthScoreColor(report.HealthScore)},
	}

	// 使用卡片式布局
	currentY := 60.0
	for i, stat := range statsData {
		var cardX float64
		if i%2 == 0 {
			cardX = 20
		} else {
			cardX = 110
		}

		if i%2 == 0 && i > 0 {
			currentY += 23
		}

		// 卡片背景
		g.pdf.SetFillColor(248, 249, 250)
		g.pdf.RectFromUpperLeftWithStyle(cardX, currentY, 85, 20, "F")

		// 标签
		g.pdf.SetY(currentY + 5)
		g.pdf.SetX(cardX + 5)
		if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
			return err
		}
		g.pdf.SetTextColor(108, 117, 125)
		if err := g.pdf.Cell(nil, stat.label); err != nil {
			return err
		}

		// 数值
		g.pdf.SetY(currentY + 12)
		g.pdf.SetX(cardX + 5)
		if err := g.pdf.SetFont("sourcehansans", "", 14); err != nil {
			return err
		}
		g.pdf.SetTextColor(uint8(stat.color[0]), uint8(stat.color[1]), uint8(stat.color[2]))
		if err := g.pdf.Cell(nil, stat.value); err != nil {
			return err
		}
	}

	// 健康度评级说明
	currentY += 30
	g.pdf.SetY(currentY)
	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(108, 117, 125)
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "健康度评级标准："); err != nil {
		return err
	}

	currentY += 8
	gradeData := []struct {
		grade  string
		range_ string
		color  []int
	}{
		{"优秀", "≥90分", []int{40, 167, 69}},
		{"良好", "70-89分", []int{13, 110, 253}},
		{"一般", "50-69分", []int{255, 193, 7}},
		{"较差", "30-49分", []int{253, 126, 20}},
		{"危险", "<30分", []int{220, 53, 69}},
	}

	if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
		return err
	}
	for _, grade := range gradeData {
		g.pdf.SetY(currentY)
		g.pdf.SetX(25)
		g.pdf.SetTextColor(uint8(grade.color[0]), uint8(grade.color[1]), uint8(grade.color[2]))
		if err := g.pdf.Cell(nil, grade.grade); err != nil {
			return err
		}
		g.pdf.SetX(50)
		g.pdf.SetTextColor(108, 117, 125)
		if err := g.pdf.Cell(nil, grade.range_); err != nil {
			return err
		}
		currentY += 6
	}

	return nil
}

// generateStatistics 生成统计分析章节（带图表）
func (g *ReportGenerator) generateStatistics(report *model.Report, defects []model.Defect) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("3. 统计分析"); err != nil {
		return err
	}

	// 3.1 缺陷类型分布
	if err := g.pdf.SetFont("sourcehansans", "", 12); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	g.pdf.SetY(60)
	if err := g.pdf.Cell(nil, "3.1 缺陷类型分布"); err != nil {
		return err
	}

	// 统计各类型缺陷数量
	typeCount := make(map[string]int)
	for _, defect := range defects {
		typeCount[defect.DefectType]++
	}

	// 生成饼图
	currentY := 70.0
	if len(typeCount) > 0 {
		chartPath := filepath.Join("reports", fmt.Sprintf("chart_pie_%d.png", time.Now().Unix()))
		if err := g.generatePieChart(typeCount, chartPath); err == nil {
			// 插入图表
			if err := g.pdf.Image(chartPath, 40, currentY, nil); err != nil {
				return err
			}
			currentY += 80
			// 删除临时图表文件
			os.Remove(chartPath)
		}
	} else {
		if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
			return err
		}
		g.pdf.SetTextColor(108, 117, 125)
		g.pdf.SetX(20)
		g.pdf.SetY(currentY)
		if err := g.pdf.Cell(nil, "当前时间范围内暂无缺陷数据，无法绘制类型分布图。"); err != nil {
			return err
		}
		currentY += 20
	}

	// 3.2 缺陷趋势分析
	currentY += 10
	g.pdf.SetY(currentY)
	if err := g.pdf.SetFont("sourcehansans", "", 12); err != nil {
		return err
	}
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "3.2 缺陷趋势分析"); err != nil {
		return err
	}

	currentY += 10
	// 按日期统计
	dateCount := make(map[string]int)
	for _, defect := range defects {
		dateStr := defect.DetectedAt.Format("2006-01-02")
		dateCount[dateStr]++
	}

	// 生成折线图
	if len(dateCount) > 0 {
		chartPath := filepath.Join("reports", fmt.Sprintf("chart_line_%d.png", time.Now().Unix()))
		if err := g.generateLineChart(dateCount, chartPath); err == nil {
			// 插入图表
			if err := g.pdf.Image(chartPath, 30, currentY, nil); err != nil {
				return err
			}
			// 删除临时图表文件
			os.Remove(chartPath)
		}
	} else {
		if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
			return err
		}
		g.pdf.SetTextColor(108, 117, 125)
		g.pdf.SetX(20)
		g.pdf.SetY(currentY)
		if err := g.pdf.Cell(nil, "当前时间范围内暂无缺陷数据，无法绘制趋势图。"); err != nil {
			return err
		}
	}

	return nil
}

// generateHighRiskDefects 生成高危缺陷列表
func (g *ReportGenerator) generateHighRiskDefects(defects []model.Defect) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("4. 高危缺陷列表"); err != nil {
		return err
	}

	// 筛选高危缺陷
	var highRiskDefects []model.Defect
	for _, defect := range defects {
		if defect.Confidence >= 0.85 || defect.Area >= 0.02 {
			highRiskDefects = append(highRiskDefects, defect)
		}
	}

	if len(highRiskDefects) == 0 {
		if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
			return err
		}
		g.pdf.SetTextColor(108, 117, 125)
		g.pdf.SetX(20)
		g.pdf.SetY(60)
		if err := g.pdf.Cell(nil, "暂无高危缺陷"); err != nil {
			return err
		}
		return nil
	}

	// 表头
	colWidths := []float64{32, 52, 24, 24, 38}
	colHeaders := []string{"缺陷类型", "位置", "面积(㎡)", "置信度", "检测时间"}
	currentY := 60.0
	if err := g.drawHighRiskHeader(currentY, colWidths, colHeaders); err != nil {
		return err
	}
	currentY += 8

	// 表格数据
	if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)

	for i, defect := range highRiskDefects {
		if currentY+7 > 278 {
			g.pdf.AddPage()
			if err := g.addSectionTitle("4. 高危缺陷列表（续）"); err != nil {
				return err
			}
			currentY = 60
			if err := g.drawHighRiskHeader(currentY, colWidths, colHeaders); err != nil {
				return err
			}
			currentY += 8
		}

		// 交替行颜色
		if i%2 == 1 {
			g.pdf.SetFillColor(248, 249, 250)
			g.pdf.RectFromUpperLeftWithStyle(20, currentY, 170, 7, "F")
		}

		x := 22.0
		g.pdf.SetY(currentY + 2)
		g.pdf.SetX(x)
		if err := g.pdf.Cell(nil, truncateTextToWidth(g.pdf, defect.DefectType, colWidths[0]-3)); err != nil {
			return err
		}
		x += colWidths[0]

		g.pdf.SetX(x)
		if err := g.pdf.Cell(nil, truncateTextToWidth(g.pdf, compactBBox(defect.BBox), colWidths[1]-3)); err != nil {
			return err
		}
		x += colWidths[1]

		g.pdf.SetX(x)
		if err := g.pdf.Cell(nil, fmt.Sprintf("%.4f", defect.Area)); err != nil {
			return err
		}
		x += colWidths[2]

		// 置信度颜色编码
		if defect.Confidence >= 0.95 {
			g.pdf.SetTextColor(220, 53, 69)
		} else if defect.Confidence >= 0.90 {
			g.pdf.SetTextColor(253, 126, 20)
		} else {
			g.pdf.SetTextColor(255, 193, 7)
		}
		g.pdf.SetX(x)
		if err := g.pdf.Cell(nil, fmt.Sprintf("%.2f%%", defect.Confidence*100)); err != nil {
			return err
		}
		g.pdf.SetTextColor(33, 37, 41)
		x += colWidths[3]

		g.pdf.SetX(x)
		if err := g.pdf.Cell(nil, defect.DetectedAt.Format("2006-01-02")); err != nil {
			return err
		}

		// 边框
		g.pdf.SetStrokeColor(222, 226, 230)
		g.pdf.RectFromUpperLeft(20, currentY, 170, 7)

		currentY += 7
	}

	return nil
}

// generateDefectDetails 生成缺陷详情（按类型分组）
func (g *ReportGenerator) generateDefectDetails(defects []model.Defect) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("5. 缺陷详细信息"); err != nil {
		return err
	}

	// 按类型分组
	typeGroups := make(map[string][]model.Defect)
	for _, defect := range defects {
		typeGroups[defect.DefectType] = append(typeGroups[defect.DefectType], defect)
	}

	// 为了保证分页稳定，先按类型名称排序
	typeNames := make([]string, 0, len(typeGroups))
	for typeName := range typeGroups {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	// 遍历每种类型
	typeIndex := 1
	currentY := 60.0

	for _, defectType := range typeNames {
		defectList := typeGroups[defectType]
		// 检查是否需要换页
		if currentY > 250 {
			g.pdf.AddPage()
			currentY = 40
		}

		// 类型标题
		if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
			return err
		}
		g.pdf.SetTextColor(33, 37, 41)
		g.pdf.SetX(20)
		g.pdf.SetY(currentY)
		if err := g.pdf.Cell(nil, fmt.Sprintf("5.%d %s (共%d个)", typeIndex, defectType, len(defectList))); err != nil {
			return err
		}
		currentY += 10

		// 缺陷列表
		if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
			return err
		}
		g.pdf.SetTextColor(73, 80, 87)

		sort.Slice(defectList, func(i, j int) bool {
			return defectList[i].DetectedAt.After(defectList[j].DetectedAt)
		})

		for i, defect := range defectList {
			// 检查是否需要换页
			if currentY > 270 {
				g.pdf.AddPage()
				currentY = 40
			}

			line1 := fmt.Sprintf("%d) 检测时间: %s   面积: %.4f㎡   置信度: %.2f%%",
				i+1, defect.DetectedAt.Format("2006-01-02"), defect.Area, defect.Confidence*100)
			g.pdf.SetY(currentY)
			g.pdf.SetX(25)
			if err := g.pdf.Cell(nil, line1); err != nil {
				return err
			}

			currentY += 5
			line2 := "位置: " + compactBBox(defect.BBox)
			nextY, err := g.drawWrappedText(25, currentY, 160, 5, line2)
			if err != nil {
				return err
			}
			currentY = nextY + 2
		}

		currentY += 5
		typeIndex++
	}

	return nil
}

// generateConclusion 生成结论与建议
func (g *ReportGenerator) generateConclusion(report *model.Report, defects []model.Defect) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("6. 结论与建议"); err != nil {
		return err
	}

	summary := buildReportSummary(report, defects)
	currentY := 60.0

	// 6.1 执行摘要
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	g.pdf.SetY(currentY)
	if err := g.pdf.Cell(nil, "6.1 执行摘要"); err != nil {
		return err
	}

	currentY += 9
	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)
	execSummary := fmt.Sprintf(
		"监测区间 %s 至 %s，共识别缺陷 %d 个，其中高危 %d 个（占比 %.1f%%）。综合健康评分 %.2f 分，健康等级 %s，风险等级 %s。",
		report.StartTime.Format("2006-01-02"),
		report.EndTime.Format("2006-01-02"),
		report.DefectCount,
		report.HighRiskCount,
		summary.HighRiskRatio*100,
		report.HealthScore,
		summary.HealthLevel,
		summary.RiskLevel,
	)
	nextY, err := g.drawWrappedText(20, currentY, 170, 6, execSummary)
	if err != nil {
		return err
	}
	currentY = nextY + 2

	// 6.2 关键发现
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	g.pdf.SetY(currentY)
	if err := g.pdf.Cell(nil, "6.2 关键发现"); err != nil {
		return err
	}
	currentY += 9

	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)
	for i, finding := range summary.KeyFindings {
		line := fmt.Sprintf("%d. %s", i+1, finding)
		nextY, err = g.drawWrappedText(20, currentY, 170, 6, line)
		if err != nil {
			return err
		}
		currentY = nextY + 1
	}

	currentY += 2

	// 6.3 风险与处置建议
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	g.pdf.SetY(currentY)
	if err := g.pdf.Cell(nil, "6.3 风险与处置建议"); err != nil {
		return err
	}
	currentY += 9

	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)
	riskText := fmt.Sprintf(
		"当前风险判定为%s。建议按“先高危、后一般”的顺序实施处置，优先闭环高危缺陷并跟踪复检结果。",
		summary.RiskLevel,
	)
	nextY, err = g.drawWrappedText(20, currentY, 170, 6, riskText)
	if err != nil {
		return err
	}
	currentY = nextY + 2

	for i, action := range summary.PriorityActions {
		line := fmt.Sprintf("%d. %s", i+1, action)
		nextY, err = g.drawWrappedText(20, currentY, 170, 6, line)
		if err != nil {
			return err
		}
		currentY = nextY + 1
	}

	currentY += 2

	// 6.4 下次检测建议
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	g.pdf.SetY(currentY)
	if err := g.pdf.Cell(nil, "6.4 下次检测建议"); err != nil {
		return err
	}
	currentY += 9

	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)
	nextInspectText := fmt.Sprintf(
		"建议下次检测时间：%s（建议间隔：%d天，依据：%s）。",
		summary.NextInspectionDate.Format("2006-01-02"),
		summary.NextInspectionIntervalDays,
		summary.NextInspectionReason,
	)
	nextY, err = g.drawWrappedText(20, currentY, 170, 6, nextInspectText)
	if err != nil {
		return err
	}
	currentY = nextY + 8

	// 报告结束标记
	if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
		return err
	}
	g.pdf.SetTextColor(173, 181, 189)
	g.pdf.SetX(20)
	g.pdf.SetY(currentY)
	if err := g.pdf.Cell(nil, "--- 报告结束 ---"); err != nil {
		return err
	}

	return nil
}

type reportSummary struct {
	HealthLevel                string
	RiskLevel                  string
	HighRiskRatio              float64
	KeyFindings                []string
	PriorityActions            []string
	NextInspectionDate         time.Time
	NextInspectionIntervalDays int
	NextInspectionReason       string
}

func buildReportSummary(report *model.Report, defects []model.Defect) reportSummary {
	summary := reportSummary{
		HealthLevel: classifyHealthLevel(report.HealthScore),
	}

	if report.DefectCount > 0 {
		summary.HighRiskRatio = float64(report.HighRiskCount) / float64(report.DefectCount)
	}

	riskLevel, days, reason := classifyRiskAndNextInspection(report.HealthScore, report.HighRiskCount, report.DefectCount)
	summary.RiskLevel = riskLevel
	summary.NextInspectionIntervalDays = days
	summary.NextInspectionReason = reason
	summary.NextInspectionDate = report.EndTime.AddDate(0, 0, days)

	typeCount := make(map[string]int)
	var totalConfidence float64
	var totalArea float64
	for _, defect := range defects {
		typeCount[defect.DefectType]++
		totalConfidence += defect.Confidence
		totalArea += defect.Area
	}

	topTypes := topDefectTypes(typeCount, 3)
	topTypesText := "无明显集中类型"
	if len(topTypes) > 0 {
		topTypesText = strings.Join(topTypes, "、")
	}

	avgConfidence := 0.0
	avgArea := 0.0
	if len(defects) > 0 {
		avgConfidence = totalConfidence / float64(len(defects))
		avgArea = totalArea / float64(len(defects))
	}

	trendText := buildTrendSummary(report, defects)

	summary.KeyFindings = []string{
		fmt.Sprintf("缺陷类型集中度：主要集中在%s。", topTypesText),
		fmt.Sprintf("缺陷趋势判断：%s。", trendText),
		fmt.Sprintf("检测结果质量：平均置信度 %.2f%%，平均缺陷面积 %.4f㎡。", avgConfidence*100, avgArea),
	}

	summary.PriorityActions = buildPriorityActions(summary.RiskLevel, report, topTypesText)
	return summary
}

func classifyHealthLevel(score float64) string {
	switch {
	case score >= 90:
		return "优秀"
	case score >= 70:
		return "良好"
	case score >= 50:
		return "一般"
	case score >= 30:
		return "较差"
	default:
		return "危险"
	}
}

func classifyRiskAndNextInspection(score float64, highRiskCount, defectCount int) (riskLevel string, days int, reason string) {
	if highRiskCount >= 5 || score < 30 {
		return "高风险", 7, "高危缺陷数量较多或健康评分过低"
	}
	if highRiskCount > 0 || score < 70 || defectCount >= 20 {
		return "中风险", 14, "存在高危缺陷或整体缺陷数量偏多"
	}
	return "低风险", 30, "高危缺陷较少且健康评分处于可控区间"
}

func topDefectTypes(typeCount map[string]int, topN int) []string {
	type pair struct {
		name  string
		count int
	}
	pairs := make([]pair, 0, len(typeCount))
	for name, count := range typeCount {
		pairs = append(pairs, pair{name: name, count: count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].name < pairs[j].name
		}
		return pairs[i].count > pairs[j].count
	})

	if topN > len(pairs) {
		topN = len(pairs)
	}
	result := make([]string, 0, topN)
	for i := 0; i < topN; i++ {
		result = append(result, fmt.Sprintf("%s(%d个)", pairs[i].name, pairs[i].count))
	}
	return result
}

func buildTrendSummary(report *model.Report, defects []model.Defect) string {
	if len(defects) < 2 {
		return "样本量不足，暂无法判断趋势"
	}

	mid := report.StartTime.Add(report.EndTime.Sub(report.StartTime) / 2)
	firstHalf := 0
	secondHalf := 0
	for _, defect := range defects {
		if defect.DetectedAt.Before(mid) {
			firstHalf++
		} else {
			secondHalf++
		}
	}

	switch {
	case firstHalf == 0 && secondHalf > 0:
		return "后半程新增明显增多，呈上升趋势"
	case secondHalf == 0 && firstHalf > 0:
		return "后半程未发现新增，呈下降趋势"
	case secondHalf > firstHalf*12/10:
		return fmt.Sprintf("后半程缺陷数(%d)高于前半程(%d)，呈上升趋势", secondHalf, firstHalf)
	case secondHalf < firstHalf*8/10:
		return fmt.Sprintf("后半程缺陷数(%d)低于前半程(%d)，呈下降趋势", secondHalf, firstHalf)
	default:
		return fmt.Sprintf("前后半程缺陷数接近（%d vs %d），整体平稳", firstHalf, secondHalf)
	}
}

func buildPriorityActions(riskLevel string, report *model.Report, topTypesText string) []string {
	actions := make([]string, 0, 5)

	if report.HighRiskCount > 0 {
		actions = append(actions, fmt.Sprintf("48小时内完成 %d 个高危缺陷复核，明确加固/修复清单。", report.HighRiskCount))
	} else {
		actions = append(actions, "本周期未识别高危缺陷，保持例行巡检频率并关注新增异常。")
	}

	actions = append(actions, fmt.Sprintf("围绕主导缺陷类型（%s）制定专项治理方案，优先处理重复出现区域。", topTypesText))

	switch riskLevel {
	case "高风险":
		actions = append(actions, "建议立即组织结构安全专项评估，必要时采取限载或临时交通管控。")
	case "中风险":
		actions = append(actions, "建议两周内完成重点区域复检，并形成闭环处置台账。")
	default:
		actions = append(actions, "建议按月开展常规复检，持续跟踪缺陷演化趋势。")
	}

	actions = append(actions, "统一记录缺陷位置、面积和复检结果，作为下周期趋势对比基线。")
	return actions
}

func (g *ReportGenerator) drawWrappedText(x, y, maxWidth, lineHeight float64, text string) (float64, error) {
	lines, err := g.wrapText(text, maxWidth)
	if err != nil {
		return y, err
	}

	for _, line := range lines {
		g.pdf.SetX(x)
		g.pdf.SetY(y)
		if err := g.pdf.Cell(nil, line); err != nil {
			return y, err
		}
		y += lineHeight
	}
	return y, nil
}

func (g *ReportGenerator) wrapText(text string, maxWidth float64) ([]string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{""}, nil
	}

	lines := make([]string, 0, 8)
	paragraphs := strings.Split(text, "\n")
	for _, paragraph := range paragraphs {
		p := strings.TrimSpace(paragraph)
		if p == "" {
			lines = append(lines, "")
			continue
		}

		current := ""
		for _, r := range p {
			candidate := current + string(r)
			width, err := g.pdf.MeasureTextWidth(candidate)
			if err != nil {
				return nil, err
			}

			if width <= maxWidth {
				current = candidate
				continue
			}

			if current == "" {
				lines = append(lines, string(r))
				continue
			}

			lines = append(lines, strings.TrimSpace(current))
			current = string(r)
		}

		if strings.TrimSpace(current) != "" {
			lines = append(lines, strings.TrimSpace(current))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines, nil
}

func countHighRiskDefects(defects []model.Defect) int {
	count := 0
	for _, defect := range defects {
		if defect.Confidence >= 0.85 || defect.Area >= 0.02 {
			count++
		}
	}
	return count
}

func compactBBox(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.ReplaceAll(clean, "\n", " ")
	clean = strings.ReplaceAll(clean, "\t", " ")
	return strings.Join(strings.Fields(clean), " ")
}

func truncateTextToWidth(pdf *gopdf.GoPdf, text string, maxWidth float64) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	width, err := pdf.MeasureTextWidth(text)
	if err == nil && width <= maxWidth {
		return text
	}

	ellipsis := "..."
	runes := []rune(text)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i]) + ellipsis
		w, e := pdf.MeasureTextWidth(candidate)
		if e == nil && w <= maxWidth {
			return candidate
		}
	}

	return ellipsis
}

func (g *ReportGenerator) drawHighRiskHeader(currentY float64, colWidths []float64, headers []string) error {
	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetFillColor(33, 37, 41)
	g.pdf.RectFromUpperLeftWithStyle(20, currentY, 170, 8, "F")
	g.pdf.SetTextColor(255, 255, 255)

	x := 22.0
	g.pdf.SetY(currentY + 2)
	for i, h := range headers {
		g.pdf.SetX(x)
		if err := g.pdf.Cell(nil, h); err != nil {
			return err
		}
		x += colWidths[i]
	}
	return nil
}

// addSectionTitle 添加章节标题
func (g *ReportGenerator) addSectionTitle(title string) error {
	if err := g.pdf.SetFont("sourcehansans", "", 16); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetFillColor(248, 249, 250)
	g.pdf.RectFromUpperLeftWithStyle(20, 40, 170, 12, "F")
	g.pdf.SetX(22)
	g.pdf.SetY(45)
	if err := g.pdf.Cell(nil, title); err != nil {
		return err
	}
	return nil
}

// generatePieChart 生成饼图
func (g *ReportGenerator) generatePieChart(data map[string]int, outputPath string) error {
	// 准备数据
	var values []chart.Value
	for label, count := range data {
		values = append(values, chart.Value{
			Label: fmt.Sprintf("%s (%d)", label, count),
			Value: float64(count),
		})
	}

	// 创建饼图
	pie := chart.PieChart{
		Width:  800,
		Height: 600,
		Values: values,
		Background: chart.Style{
			FillColor: drawing.ColorFromAlphaMixedRGBA(255, 255, 255, 255),
		},
	}

	// 保存到文件
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return pie.Render(chart.PNG, file)
}

// generateLineChart 生成折线图
func (g *ReportGenerator) generateLineChart(data map[string]int, outputPath string) error {
	// 准备X轴和Y轴数据
	var xValues []time.Time
	var yValues []float64

	// 将map转换为有序切片
	type dateCount struct {
		date  time.Time
		count int
	}
	var sortedData []dateCount

	for dateStr, count := range data {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		sortedData = append(sortedData, dateCount{date: date, count: count})
	}

	// 按日期排序
	for i := 0; i < len(sortedData)-1; i++ {
		for j := i + 1; j < len(sortedData); j++ {
			if sortedData[i].date.After(sortedData[j].date) {
				sortedData[i], sortedData[j] = sortedData[j], sortedData[i]
			}
		}
	}

	// 提取数据
	for _, item := range sortedData {
		xValues = append(xValues, item.date)
		yValues = append(yValues, float64(item.count))
	}

	// 创建折线图
	graph := chart.Chart{
		Width:  1000,
		Height: 600,
		Background: chart.Style{
			FillColor: drawing.ColorFromAlphaMixedRGBA(255, 255, 255, 255),
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "缺陷数量",
				XValues: xValues,
				YValues: yValues,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromAlphaMixedRGBA(13, 110, 253, 255),
					StrokeWidth: 2,
				},
			},
		},
	}

	// 保存到文件
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return graph.Render(chart.PNG, file)
}

// getHealthScoreColor 根据健康度评分获取颜色
func getHealthScoreColor(score float64) []int {
	if score >= 90 {
		return []int{40, 167, 69} // 绿色（优秀）
	} else if score >= 70 {
		return []int{13, 110, 253} // 蓝色（良好）
	} else if score >= 50 {
		return []int{255, 193, 7} // 黄色（一般）
	} else if score >= 30 {
		return []int{253, 126, 20} // 橙色（较差）
	}
	return []int{220, 53, 69} // 红色（危险）
}
