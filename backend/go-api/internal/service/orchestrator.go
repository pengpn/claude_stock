package service

import (
	"context"
	"fmt"
	"log"
	"stock-analysis-api/backend/go-api/internal/client"
	"stock-analysis-api/backend/go-api/internal/llm"
	"stock-analysis-api/backend/go-api/internal/model"
)

// SSEEvent SSE事件
type SSEEvent struct {
	Event string
	Data  interface{}
}

// AnalysisOrchestrator 分析编排器
type AnalysisOrchestrator struct {
	pythonClient *client.PythonClient
	llmClient    llm.LLMClient
}

func NewAnalysisOrchestrator(pythonClient *client.PythonClient, llmClient llm.LLMClient) *AnalysisOrchestrator {
	return &AnalysisOrchestrator{
		pythonClient: pythonClient,
		llmClient:    llmClient,
	}
}

// Analyze 执行完整分析流程
func (ao *AnalysisOrchestrator) Analyze(ctx context.Context, code string, eventChan chan<- SSEEvent) error {
	defer close(eventChan)

	// 步骤0: 获取Python分析数据
	eventChan <- SSEEvent{
		Event: "progress",
		Data: map[string]interface{}{
			"step":     "fetching_data",
			"message":  "正在获取股票数据...",
			"progress": 10,
		},
	}

	pythonData, err := ao.pythonClient.Analyze(code)
	if err != nil {
		return fmt.Errorf("获取数据失败: %w", err)
	}

	// 准备LLM输入数据
	llmData := ao.prepareLLMData(pythonData)

	// 存储各步骤结果
	results := make(map[string]string)

	// 步骤1: 综合分析
	if err := ao.runStep(ctx, llm.StepComprehensive, "综合分析", llmData, results, eventChan, 20); err != nil {
		return err
	}
	llmData["comprehensive_analysis"] = results[string(llm.StepComprehensive)]

	// 步骤2: 多头观点
	if err := ao.runStep(ctx, llm.StepDebateBull, "多头观点", llmData, results, eventChan, 40); err != nil {
		return err
	}
	llmData["bull_case"] = results[string(llm.StepDebateBull)]

	// 步骤3: 空头观点
	if err := ao.runStep(ctx, llm.StepDebateBear, "空头观点", llmData, results, eventChan, 60); err != nil {
		return err
	}
	llmData["bear_case"] = results[string(llm.StepDebateBear)]

	// 步骤4: 交易员决策
	if err := ao.runStep(ctx, llm.StepTrader, "交易员决策", llmData, results, eventChan, 80); err != nil {
		return err
	}
	llmData["trader_decision"] = results[string(llm.StepTrader)]

	// 步骤5: 最终决策
	if err := ao.runStep(ctx, llm.StepFinal, "最终决策", llmData, results, eventChan, 100); err != nil {
		return err
	}

	// 发送完成事件
	eventChan <- SSEEvent{
		Event: "done",
		Data:  map[string]string{"message": "分析完成"},
	}

	return nil
}

func (ao *AnalysisOrchestrator) runStep(
	ctx context.Context,
	step llm.AnalysisStep,
	stepName string,
	data map[string]interface{},
	results map[string]string,
	eventChan chan<- SSEEvent,
	progress int,
) error {
	log.Printf("开始执行: %s", stepName)

	var content string
	callback := func(delta string) error {
		content += delta
		// 发送流式内容
		eventChan <- SSEEvent{
			Event: "analysis_step",
			Data: map[string]interface{}{
				"step":     string(step),
				"role":     stepName,
				"content":  delta,
				"progress": progress,
			},
		}
		return nil
	}

	if err := ao.llmClient.StreamAnalyze(ctx, step, data, callback); err != nil {
		return fmt.Errorf("%s失败: %w", stepName, err)
	}

	results[string(step)] = content
	log.Printf("完成执行: %s", stepName)
	return nil
}

func (ao *AnalysisOrchestrator) prepareLLMData(pythonData *model.PythonAnalysisResponse) map[string]interface{} {
	return map[string]interface{}{
		"code":            pythonData.Code,
		"name":            pythonData.Name,
		"industry":        pythonData.BasicInfo.Industry,
		"market_cap":      pythonData.BasicInfo.MarketCap,
		"pe_ttm":          pythonData.BasicInfo.PETTM,
		"pb":              pythonData.BasicInfo.PB,
		"latest_price":    pythonData.Price.LatestPrice,
		"roe":             pythonData.FinancialMetrics.ROE,
		"debt_ratio":      pythonData.FinancialMetrics.DebtRatio,
		"revenue_growth":  pythonData.FinancialMetrics.RevenueGrowth,
		"profit_growth":   pythonData.FinancialMetrics.ProfitGrowth,
		"risks":           pythonData.Risks,
	}
}
