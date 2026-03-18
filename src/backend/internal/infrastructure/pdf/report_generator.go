// Package pdf 实现PDF报表生成功能
package pdf

import (
	"fmt"
	"os"
	"path/filepath"
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
	if err := g.generateStatistics(report, defects); err != nil {
		return fmt.Errorf("生成统计分析失败: %v", err)
	}
	if err := g.generateHighRiskDefects(defects); err != nil {
		return fmt.Errorf("生成高危缺陷列表失败: %v", err)
	}
	if err := g.generateDefectDetails(defects); err != nil {
		return fmt.Errorf("生成缺陷详情失败: %v", err)
	}
	if err := g.generateConclusion(report); err != nil {
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
	// 尝试加载文泉驿微米黑字体（TTC格式，需要特殊处理）
	ttcPath := filepath.Join(filepath.Dir(g.fontPath), "wqy-microhei.ttc")
	if _, err := os.Stat(ttcPath); err == nil {
		// TTC文件存在，尝试读取并加载第一个字体
		data, err := os.ReadFile(ttcPath)
		if err != nil {
			return fmt.Errorf("读取TTC字体文件失败: %v", err)
		}
		// TTC文件格式：前4字节是"ttcf"标识，然后是版本号，然后是字体数量
		// 跳过TTC头部，直接提取第一个TTF字体
		// 这是一个简化处理，实际TTC格式更复杂
		if len(data) < 12 {
			return fmt.Errorf("TTC文件格式无效")
		}
		// 直接使用字体数据（gopdf可能不支持TTC）
		if err := g.pdf.AddTTFFontData("sourcehansans", data); err != nil {
			// TTC加载失败，尝试使用系统Liberation字体作为fallback
			fallbackPath := "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf"
			if err2 := g.pdf.AddTTFFont("sourcehansans", fallbackPath); err2 != nil {
				return fmt.Errorf("加载字体失败: TTC错误=%v, Fallback错误=%v", err, err2)
			}
		}
	} else {
		// 检查原始字体文件是否存在
		if _, err := os.Stat(g.fontPath); os.IsNotExist(err) {
			// 尝试使用系统Liberation字体
			fallbackPath := "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf"
			if err := g.pdf.AddTTFFont("sourcehansans", fallbackPath); err != nil {
				return fmt.Errorf("字体文件不存在且fallback失败: %s, %v", g.fontPath, err)
			}
		} else {
			// 添加TTF字体（gopdf原生支持）
			if err := g.pdf.AddTTFFont("sourcehansans", g.fontPath); err != nil {
				// 如果加载失败，尝试使用系统Liberation字体
				fallbackPath := "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf"
				if err2 := g.pdf.AddTTFFont("sourcehansans", fallbackPath); err2 != nil {
					return fmt.Errorf("加载字体文件失败: 原始错误=%v, Fallback错误=%v", err, err2)
				}
			}
		}
	}

	// 设置默认字体
	if err := g.pdf.SetFont("sourcehansans", "", 12); err != nil {
		return fmt.Errorf("设置字体失败: %v", err)
	}

	return nil
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
	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	currentY := 60.0

	// 表头背景
	g.pdf.SetFillColor(33, 37, 41)
	g.pdf.RectFromUpperLeftWithStyle(20, currentY, 170, 8, "F")

	// 表头文字
	g.pdf.SetTextColor(255, 255, 255)
	headers := []struct {
		text string
		x    float64
	}{
		{"缺陷类型", 22},
		{"位置", 50},
		{"面积(㎡)", 100},
		{"置信度", 125},
		{"检测时间", 150},
	}

	g.pdf.SetY(currentY + 2)
	for _, h := range headers {
		g.pdf.SetX(h.x)
		if err := g.pdf.Cell(nil, h.text); err != nil {
			return err
		}
	}

	currentY += 8

	// 表格数据
	if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)

	for i, defect := range highRiskDefects {
		// 交替行颜色
		if i%2 == 1 {
			g.pdf.SetFillColor(248, 249, 250)
			g.pdf.RectFromUpperLeftWithStyle(20, currentY, 170, 7, "F")
		}

		g.pdf.SetY(currentY + 2)
		g.pdf.SetX(22)
		if err := g.pdf.Cell(nil, defect.DefectType); err != nil {
			return err
		}

		g.pdf.SetX(50)
		if err := g.pdf.Cell(nil, defect.BBox); err != nil {
			return err
		}

		g.pdf.SetX(100)
		if err := g.pdf.Cell(nil, fmt.Sprintf("%.4f", defect.Area)); err != nil {
			return err
		}

		// 置信度颜色编码
		if defect.Confidence >= 0.95 {
			g.pdf.SetTextColor(220, 53, 69)
		} else if defect.Confidence >= 0.90 {
			g.pdf.SetTextColor(253, 126, 20)
		} else {
			g.pdf.SetTextColor(255, 193, 7)
		}
		g.pdf.SetX(125)
		if err := g.pdf.Cell(nil, fmt.Sprintf("%.2f%%", defect.Confidence*100)); err != nil {
			return err
		}
		g.pdf.SetTextColor(33, 37, 41)

		g.pdf.SetX(150)
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

	// 遍历每种类型
	typeIndex := 1
	currentY := 60.0

	for defectType, defectList := range typeGroups {
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

		for i, defect := range defectList {
			// 检查是否需要换页
			if currentY > 270 {
				g.pdf.AddPage()
				currentY = 40
			}

			g.pdf.SetY(currentY)
			g.pdf.SetX(25)
			text := fmt.Sprintf("%d. 边界框:%s 面积:%.4f㎡ 置信度:%.2f%% 检测时间:%s",
				i+1, defect.BBox, defect.Area, defect.Confidence*100,
				defect.DetectedAt.Format("2006-01-02"))
			if err := g.pdf.Cell(nil, text); err != nil {
				return err
			}
			currentY += 7
		}

		currentY += 5
		typeIndex++
	}

	return nil
}

// generateConclusion 生成结论与建议
func (g *ReportGenerator) generateConclusion(report *model.Report) error {
	g.pdf.AddPage()

	// 章节标题
	if err := g.addSectionTitle("6. 结论与建议"); err != nil {
		return err
	}

	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)

	// 整体评估
	g.pdf.SetX(20)
	g.pdf.SetY(60)
	if err := g.pdf.Cell(nil, "6.1 整体评估"); err != nil {
		return err
	}

	currentY := 70.0
	assessment := ""
	if report.HealthScore >= 90 {
		assessment = "桥梁整体状况优秀，结构完好，无明显缺陷。建议按常规周期进行检测维护。"
	} else if report.HealthScore >= 70 {
		assessment = "桥梁整体状况良好，存在少量轻微缺陷。建议加强日常巡查，定期监测缺陷发展情况。"
	} else if report.HealthScore >= 50 {
		assessment = "桥梁整体状况一般，存在一定数量的缺陷。建议尽快安排专业检测，制定维修方案。"
	} else if report.HealthScore >= 30 {
		assessment = "桥梁整体状况较差，存在较多缺陷，部分为高危缺陷。建议立即开展详细检测，制定加固维修方案。"
	} else {
		assessment = "桥梁整体状况危险，存在大量高危缺陷，可能影响结构安全。建议立即采取临时加固措施，限制通行，尽快开展抢修。"
	}

	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)
	g.pdf.SetX(20)
	g.pdf.SetY(currentY)
	if err := g.pdf.Cell(nil, assessment); err != nil {
		return err
	}

	// 维护建议
	currentY += 20
	g.pdf.SetY(currentY)
	if err := g.pdf.SetFont("sourcehansans", "", 11); err != nil {
		return err
	}
	g.pdf.SetTextColor(33, 37, 41)
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "6.2 维护建议"); err != nil {
		return err
	}

	currentY += 10
	suggestions := []string{
		"1. 对检测到的高危缺陷进行重点监测，记录发展趋势",
		"2. 制定针对性的维修计划，优先处理高危缺陷",
		"3. 加强桥梁日常巡查频率，及时发现新增缺陷",
		"4. 建立缺陷档案，跟踪缺陷发展历史",
		"5. 定期开展无人机智能检测，提高检测效率和覆盖率",
	}

	if err := g.pdf.SetFont("sourcehansans", "", 10); err != nil {
		return err
	}
	g.pdf.SetTextColor(73, 80, 87)
	for _, suggestion := range suggestions {
		g.pdf.SetY(currentY)
		g.pdf.SetX(20)
		if err := g.pdf.Cell(nil, suggestion); err != nil {
			return err
		}
		currentY += 8
	}

	// 报告结束标记
	currentY += 15
	g.pdf.SetY(currentY)
	if err := g.pdf.SetFont("sourcehansans", "", 9); err != nil {
		return err
	}
	g.pdf.SetTextColor(173, 181, 189)
	g.pdf.SetX(20)
	if err := g.pdf.Cell(nil, "--- 报告结束 ---"); err != nil {
		return err
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
