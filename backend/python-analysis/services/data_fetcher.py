import akshare as ak
import pandas as pd
import time
from typing import Optional, Dict, Any
from utils.logger import logger


class StockDataFetcher:
    """股票数据获取服务
    
    使用可靠的 akshare API：
    - stock_individual_info_em: 基本信息（名称、行业、市值）
    - stock_bid_ask_em: 实时行情（最新价、涨跌幅）
    - stock_financial_abstract_ths: 财务摘要（ROE、负债率、增长率、EPS、每股净资产）
    """

    def __init__(self):
        self.logger = logger
        self.request_interval = 1.0  # 请求间隔（秒），避免被限流
        self.max_retries = 3

    def _safe_float(self, value) -> Optional[float]:
        """安全转换为浮点数"""
        if value is None or value == '' or value == '--' or value is False:
            return None
        try:
            if pd.isna(value):
                return None
            s = str(value).replace('%', '').replace(',', '').replace('亿', '').replace('万', '')
            return float(s)
        except (ValueError, TypeError):
            return None

    def _retry_call(self, func, description: str):
        """带重试的 API 调用"""
        for attempt in range(1, self.max_retries + 1):
            try:
                return func()
            except Exception as e:
                self.logger.warning(f"{description} 第{attempt}次尝试失败: {e}")
                if attempt < self.max_retries:
                    wait_time = self.request_interval * attempt
                    self.logger.info(f"等待 {wait_time}s 后重试...")
                    time.sleep(wait_time)
                else:
                    raise

    def get_basic_info(self, code: str) -> Dict[str, Any]:
        """获取股票基本信息（stock_individual_info_em）"""
        try:
            df = self._retry_call(
                lambda: ak.stock_individual_info_em(symbol=code),
                "获取基本信息"
            )
            info = {}
            for _, row in df.iterrows():
                info[row['item']] = row['value']

            return {
                "code": code,
                "name": info.get("股票简称", ""),
                "industry": info.get("行业", ""),
                "market_cap": self._safe_float(info.get("总市值")),
                "total_shares": self._safe_float(info.get("总股本")),
                "listing_date": info.get("上市时间", ""),
            }
        except Exception as e:
            self.logger.error(f"获取基本信息失败: {e}")
            return {"code": code, "error": str(e)}

    def get_realtime_price(self, code: str) -> Dict[str, Any]:
        """获取实时行情（stock_bid_ask_em）"""
        try:
            df = self._retry_call(
                lambda: ak.stock_bid_ask_em(symbol=code),
                "获取实时行情"
            )
            info = dict(zip(df['item'], df['value']))

            return {
                "latest_price": self._safe_float(info.get("最新")),
                "price_change_pct": self._safe_float(info.get("涨幅")),
                "date": pd.Timestamp.now().strftime("%Y-%m-%d"),
            }
        except Exception as e:
            self.logger.error(f"获取实时行情失败: {e}")
            return {"error": str(e)}

    def get_financial_summary(self, code: str) -> Dict[str, Any]:
        """获取财务摘要（stock_financial_abstract_ths）
        
        返回最新一期的财务指标，包括：ROE、毛利率、净利率、负债率、
        流动比率、营收增长率、净利润增长率、EPS、每股净资产等
        """
        try:
            df = self._retry_call(
                lambda: ak.stock_financial_abstract_ths(symbol=code),
                "获取财务摘要"
            )
            if df is None or df.empty:
                return {"error": "无财务数据"}

            # 按报告期降序排列，取最新一期
            df_sorted = df.sort_values('报告期', ascending=False)
            latest = df_sorted.iloc[0]

            return {
                "report_date": str(latest.get("报告期", "")),
                "roe": self._safe_float(latest.get("净资产收益率")),
                "gross_margin": self._safe_float(latest.get("销售毛利率")),
                "net_margin": self._safe_float(latest.get("销售净利率")),
                "debt_ratio": self._safe_float(latest.get("资产负债率")),
                "current_ratio": self._safe_float(latest.get("流动比率")),
                "quick_ratio": self._safe_float(latest.get("速动比率")),
                "revenue_growth": self._safe_float(latest.get("营业总收入同比增长率")),
                "profit_growth": self._safe_float(latest.get("净利润同比增长率")),
                "eps": self._safe_float(latest.get("基本每股收益")),
                "nav_per_share": self._safe_float(latest.get("每股净资产")),
                "net_profit": str(latest.get("净利润", "")),
                "revenue": str(latest.get("营业总收入", "")),
            }
        except Exception as e:
            self.logger.error(f"获取财务摘要失败: {e}")
            return {"error": str(e)}

    def _compute_pe_pb(self, price: float, eps: Optional[float],
                       nav_per_share: Optional[float],
                       report_date: str) -> Dict[str, Optional[float]]:
        """根据 EPS 和每股净资产计算 PE_TTM 和 PB"""
        pe_ttm = None
        pb = None

        if price and eps and eps > 0:
            # 根据报告期推算年化 EPS
            # 例如：Q3 报告（09-30），EPS 是 9 个月的，年化 = eps * 4/3
            annualized_eps = eps
            if report_date:
                if "09-30" in report_date or "9-30" in report_date:
                    annualized_eps = eps * 4 / 3
                elif "06-30" in report_date or "6-30" in report_date:
                    annualized_eps = eps * 2
                elif "03-31" in report_date or "3-31" in report_date:
                    annualized_eps = eps * 4
                # 12-31 的就是全年，无需调整

            pe_ttm = round(price / annualized_eps, 2)

        if price and nav_per_share and nav_per_share > 0:
            pb = round(price / nav_per_share, 2)

        return {"pe_ttm": pe_ttm, "pb": pb}

    def fetch_all(self, code: str) -> Dict[str, Any]:
        """获取所有数据"""
        self.logger.info(f"开始获取股票数据: {code}")

        # 1. 基本信息
        basic_info = self.get_basic_info(code)
        time.sleep(self.request_interval)

        # 2. 实时行情
        price_data = self.get_realtime_price(code)
        time.sleep(self.request_interval)

        # 3. 财务摘要
        financial = self.get_financial_summary(code)

        # 4. 根据财务数据计算 PE/PB，补充到 basic_info
        latest_price = price_data.get("latest_price")
        if latest_price and "error" not in financial:
            valuation = self._compute_pe_pb(
                price=latest_price,
                eps=financial.get("eps"),
                nav_per_share=financial.get("nav_per_share"),
                report_date=financial.get("report_date", ""),
            )
            basic_info["pe_ttm"] = valuation["pe_ttm"]
            basic_info["pb"] = valuation["pb"]

        result = {
            "code": code,
            "basic_info": basic_info,
            "price": price_data,
            "financial_summary": financial,
        }

        self.logger.info(
            f"数据获取完成: {code}, "
            f"价格={price_data.get('latest_price')}, "
            f"PE={basic_info.get('pe_ttm')}, "
            f"PB={basic_info.get('pb')}, "
            f"ROE={financial.get('roe')}"
        )
        return result
