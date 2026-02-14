from typing import Dict, Any, List
from utils.logger import logger


class FinancialAnalyzer:
    """财务分析服务
    
    基于 stock_financial_abstract_ths 返回的财务摘要进行分析
    """

    def __init__(self):
        self.logger = logger

    def _safe_num(self, value, default=0.0) -> float:
        """安全获取数值"""
        if value is None:
            return default
        try:
            return float(value)
        except (ValueError, TypeError):
            return default

    def extract_metrics(self, financial: Dict[str, Any]) -> Dict[str, Any]:
        """从财务摘要中提取标准化指标"""
        return {
            "roe": self._safe_num(financial.get("roe")),
            "roa": 0.0,  # stock_financial_abstract_ths 不提供 ROA
            "gross_margin": self._safe_num(financial.get("gross_margin")),
            "net_margin": self._safe_num(financial.get("net_margin")),
            "debt_ratio": self._safe_num(financial.get("debt_ratio")),
            "current_ratio": self._safe_num(financial.get("current_ratio")),
            "revenue_growth": self._safe_num(financial.get("revenue_growth")),
            "profit_growth": self._safe_num(financial.get("profit_growth")),
        }

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
        if roe and roe < 5:
            risks.append("净资产收益率较低")

        # 流动性风险
        current_ratio = metrics.get("current_ratio", 0)
        if current_ratio and current_ratio < 1:
            risks.append("流动比率低于1，短期偿债能力弱")

        # 估值风险
        pe = basic_info.get("pe_ttm")
        if pe and pe > 50:
            risks.append("市盈率较高，估值偏贵")

        # 增长放缓
        revenue_growth = metrics.get("revenue_growth", 0)
        profit_growth = metrics.get("profit_growth", 0)
        if revenue_growth is not None and revenue_growth < 0:
            risks.append("营业收入同比下降")
        if profit_growth is not None and profit_growth < 0:
            risks.append("净利润同比下降")

        return risks if risks else ["未检测到明显风险"]

    def analyze(self, stock_data: Dict[str, Any]) -> Dict[str, Any]:
        """综合分析"""
        self.logger.info(f"开始财务分析: {stock_data.get('code')}")

        financial = stock_data.get("financial_summary", {})
        basic_info = stock_data.get("basic_info", {})

        if "error" in financial:
            self.logger.warning(f"财务数据获取失败，使用默认值: {financial.get('error')}")
            financial = {}

        metrics = self.extract_metrics(financial)
        risks = self.detect_risks(metrics, basic_info)

        return {
            "financial_metrics": metrics,
            "risks": risks,
        }
