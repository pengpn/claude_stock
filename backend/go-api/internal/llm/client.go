package llm

import "context"

// AnalysisStep 分析步骤类型
type AnalysisStep string

const (
	StepComprehensive AnalysisStep = "comprehensive"
	StepDebateBull    AnalysisStep = "debate_bull"
	StepDebateBear    AnalysisStep = "debate_bear"
	StepTrader        AnalysisStep = "trader"
	StepFinal         AnalysisStep = "final"
)

// StreamCallback 流式响应回调
type StreamCallback func(content string) error

// LLMClient LLM客户端接口
type LLMClient interface {
	// StreamAnalyze 流式分析
	StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error
}
