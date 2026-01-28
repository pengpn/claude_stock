package llm

import "fmt"

// GetSystemPrompt 获取系统提示词
func GetSystemPrompt(step AnalysisStep) string {
	prompts := map[AnalysisStep]string{
		StepComprehensive: `你是一位资深的A股投资分析师，擅长客观中立地分析上市公司。请基于提供的财务数据进行综合分析，包括：
1. 公司基本情况和行业地位
2. 财务健康度（盈利能力、偿债能力）
3. 估值水平评估
4. 主要风险点

要求：客观中立，基于数据，200-300字，结构清晰。`,

		StepDebateBull: `你是一位乐观的多头投资者，擅长挖掘股票的投资价值和上涨潜力。请从多头角度分析：
1. 最吸引人的3-5个投资亮点
2. 为什么现在是好的买入时机
3. 未来的上涨驱动力

要求：积极正面但基于数据，150-200字，突出投资价值。`,

		StepDebateBear: `你是一位谨慎的空头投资者，擅长识别风险和质疑过度乐观的预期。请从空头角度分析：
1. 最大的3-5个风险点
2. 为什么当前估值可能不便宜
3. 哪些因素可能导致下跌

要求：批判谨慎但基于逻辑，150-200字，突出风险因素。`,

		StepTrader: `你是一位实战经验丰富的A股交易员，擅长将分析转化为具体的交易决策。基于前面的多空分析，给出：
1. 操作方向（买入/持有/卖出）
2. 建议仓位（轻仓5-10%/中仓10-20%/重仓20%+）
3. 参考买入价位区间
4. 止损位设置
5. 预期持有周期

要求：具体可执行，考虑风险收益比，150-200字。`,

		StepFinal: `你是投资决策委员会的风险管理官，负责综合各方意见给出最终决策。请提供：
1. 风险等级评估（高/中/低风险）
2. 综合投资建议（买入/持有/卖出）
3. 信心指数（0-100）
4. 决策理由总结

要求：平衡风险和收益，给出明确结论，200-250字。`,
	}

	return prompts[step]
}

// BuildUserPrompt 构建用户提示词
func BuildUserPrompt(step AnalysisStep, data map[string]interface{}) string {
	name := data["name"].(string)
	code := data["code"].(string)

	switch step {
	case StepComprehensive:
		return fmt.Sprintf(`请分析【%s(%s)】：

【基本信息】
- 行业: %v
- 市值: %.2f亿元
- 最新价: %.2f元
- PE: %.2f, PB: %.2f

【财务指标】
- ROE: %.2f%%
- 资产负债率: %.2f%%
- 营收增长: %.2f%%
- 净利润增长: %.2f%%

【风险信号】
%v

请进行综合分析。`,
			name, code,
			data["industry"],
			data["market_cap"],
			data["latest_price"],
			data["pe_ttm"],
			data["pb"],
			data["roe"],
			data["debt_ratio"],
			data["revenue_growth"],
			data["profit_growth"],
			data["risks"])

	case StepDebateBull, StepDebateBear:
		previous := data["comprehensive_analysis"].(string)
		return fmt.Sprintf(`基于以下综合分析，请给出【%s】的看%s观点：

【综合分析】
%s

【关键数据】
- ROE: %.2f%%
- 资产负债率: %.2f%%
- 营收增长: %.2f%%

请从%s角度分析。`,
			name,
			map[AnalysisStep]string{StepDebateBull: "多", StepDebateBear: "空"}[step],
			previous,
			data["roe"],
			data["debt_ratio"],
			data["revenue_growth"],
			map[AnalysisStep]string{StepDebateBull: "多头", StepDebateBear: "空头"}[step])

	case StepTrader:
		return fmt.Sprintf(`基于以下分析，给出【%s】的交易建议：

【综合分析】
%s

【多头观点】
%s

【空头观点】
%s

【当前价格】%.2f元

请给出具体的交易建议。`,
			name,
			data["comprehensive_analysis"],
			data["bull_case"],
			data["bear_case"],
			data["latest_price"])

	case StepFinal:
		return fmt.Sprintf(`基于完整分析链，给出【%s】的最终投资建议：

【综合分析】
%s

【多头观点】
%s

【空头观点】
%s

【交易员建议】
%s

请给出最终决策（包含：风险等级、投资建议、信心指数、理由）。`,
			name,
			data["comprehensive_analysis"],
			data["bull_case"],
			data["bear_case"],
			data["trader_decision"])

	default:
		return "请进行分析"
	}
}
