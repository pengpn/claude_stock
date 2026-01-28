package model

// StockAnalyzeRequest 股票分析请求
type StockAnalyzeRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name"`
}

// BasicInfo 基本信息
type BasicInfo struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Industry   string  `json:"industry"`
	MarketCap  float64 `json:"market_cap"`
	PETTM      float64 `json:"pe_ttm"`
	PB         float64 `json:"pb"`
}

// PriceInfo 价格信息
type PriceInfo struct {
	LatestPrice     float64 `json:"latest_price"`
	PriceChangePct  float64 `json:"price_change_pct"`
	Date            string  `json:"date"`
}

// FinancialMetrics 财务指标
type FinancialMetrics struct {
	ROE            float64 `json:"roe"`
	ROA            float64 `json:"roa"`
	GrossMargin    float64 `json:"gross_margin"`
	NetMargin      float64 `json:"net_margin"`
	DebtRatio      float64 `json:"debt_ratio"`
	CurrentRatio   float64 `json:"current_ratio"`
	RevenueGrowth  float64 `json:"revenue_growth"`
	ProfitGrowth   float64 `json:"profit_growth"`
}

// PythonAnalysisResponse Python分析响应
type PythonAnalysisResponse struct {
	Code              string           `json:"code"`
	Name              string           `json:"name"`
	BasicInfo         BasicInfo        `json:"basic_info"`
	Price             PriceInfo        `json:"price"`
	FinancialMetrics  FinancialMetrics `json:"financial_metrics"`
	Risks             []string         `json:"risks"`
}
