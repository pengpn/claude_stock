from typing import Dict, Any, List
from utils.logger import logger

class FinancialAnalyzer:
    """财务分析服务"""

    def __init__(self):
        self.logger = logger

    def extract_latest_metrics(self, indicators: list) -> Dict[str, Any]:
        """提取最新财务指标"""
        if not indicators:
            return {}

        latest = indicators[0]

        return {
            "roe": self._safe_get(latest, "净资产收益率", "ROE"),
            "roa": self._safe_get(latest, "总资产净利率", "ROA"),
            "gross_margin": self._safe_get(latest, "销售毛利率", "毛利率"),
            "net_margin": self._safe_get(latest, "销售净利率", "净利率"),
            "debt_ratio": self._safe_get(latest, "资产负债率", "负债率"),
            "current_ratio": self._safe_get(latest, "流动比率"),
            "quick_ratio": self._safe_get(latest, "速动比率"),
        }

    def _safe_get(self, data: dict, *keys) -> float:
        """安全获取字典值"""
        for key in keys:
            if key in data:
                val = data[key]
                if val and val != '--':
                    try:
                        return float(str(val).replace('%', ''))
                    except:
                        pass
        return 0.0

    def calculate_growth(self, indicators: list) -> Dict[str, float]:
        """计算增长率"""
        if len(indicators) < 2:
            return {"revenue_growth": 0.0, "profit_growth": 0.0}

        latest = indicators[0]
        previous = indicators[1]

        revenue_growth = self._calc_growth_rate(
            self._safe_get(latest, "营业总收入"),
            self._safe_get(previous, "营业总收入")
        )

        profit_growth = self._calc_growth_rate(
            self._safe_get(latest, "净利润"),
            self._safe_get(previous, "净利润")
        )

        return {
            "revenue_growth": revenue_growth,
            "profit_growth": profit_growth
        }

    def _calc_growth_rate(self, current: float, previous: float) -> float:
        """计算增长率"""
        if previous == 0 or previous is None:
            return 0.0
        return ((current - previous) / previous) * 100

    def detect_risks(self, metrics: Dict[str, Any], basic_info: Dict[str, Any]) -> List[str]:
        """检测财务风险"""
        risks = []

        # 高负债风险
        debt_ratio = metrics.get("debt_ratio", 0)
        if debt_ratio > 70:
            risks.append("资产负债率过高")
        elif debt_ratio > 60:
            risks.append("资产负债率偏高")

        # 低盈利能力
        roe = metrics.get("roe", 0)
        if roe < 5:
            risks.append("净资产收益率较低")

        # 流动性风险
        current_ratio = metrics.get("current_ratio", 0)
        if current_ratio < 1:
            risks.append("流动比率低于1，短期偿债能力弱")

        # 估值风险
        pe = basic_info.get("pe_ttm", 0)
        if pe and pe > 50:
            risks.append("市盈率较高，估值偏贵")

        return risks if risks else ["未检测到明显风险"]

    def analyze(self, stock_data: Dict[str, Any]) -> Dict[str, Any]:
        """综合分析"""
        self.logger.info(f"开始财务分析: {stock_data.get('code')}")

        indicators = stock_data.get("financial_indicators", [])
        basic_info = stock_data.get("basic_info", {})

        metrics = self.extract_latest_metrics(indicators)
        growth = self.calculate_growth(indicators)
        risks = self.detect_risks(metrics, basic_info)

        return {
            "financial_metrics": {**metrics, **growth},
            "risks": risks
        }
