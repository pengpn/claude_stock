import akshare as ak
import pandas as pd
from typing import Optional, Dict, Any
from utils.logger import logger

class StockDataFetcher:
    """股票数据获取服务"""

    def __init__(self):
        self.logger = logger

    def safe_float(self, value) -> Optional[float]:
        """安全转换为浮点数"""
        if value is None or value == '' or value == '--':
            return None
        try:
            if pd.isna(value):
                return None
            if isinstance(value, str):
                value = value.replace('%', '').replace(',', '').replace('亿', '')
            return float(value)
        except (ValueError, TypeError):
            return None

    def get_basic_info(self, code: str) -> Dict[str, Any]:
        """获取股票基本信息"""
        try:
            df = ak.stock_individual_info_em(symbol=code)
            info = {}
            for _, row in df.iterrows():
                info[row['item']] = row['value']

            return {
                "code": code,
                "name": info.get("股票简称", ""),
                "industry": info.get("行业", ""),
                "market_cap": self.safe_float(info.get("总市值")),
                "pe_ttm": self.safe_float(info.get("市盈率(动态)")),
                "pb": self.safe_float(info.get("市净率")),
                "listing_date": info.get("上市时间", "")
            }
        except Exception as e:
            self.logger.error(f"获取基本信息失败: {e}")
            return {"code": code, "error": str(e)}

    def get_latest_price(self, code: str) -> Dict[str, Any]:
        """获取最新价格数据"""
        try:
            df = ak.stock_zh_a_hist(
                symbol=code,
                period="daily",
                adjust="qfq"
            )
            if df is None or df.empty:
                return {"error": "无价格数据"}

            latest = df.iloc[-1]
            return {
                "latest_price": self.safe_float(latest['收盘']),
                "price_change_pct": self.safe_float(latest['涨跌幅']),
                "date": str(latest['日期'])
            }
        except Exception as e:
            self.logger.error(f"获取价格数据失败: {e}")
            return {"error": str(e)}

    def get_financial_indicators(self, code: str, limit: int = 4) -> list:
        """获取财务指标"""
        try:
            df = ak.stock_financial_analysis_indicator(symbol=code)
            if df is None or df.empty:
                return []
            return df.head(limit).to_dict(orient='records')
        except Exception as e:
            self.logger.error(f"获取财务指标失败: {e}")
            return []

    def fetch_all(self, code: str) -> Dict[str, Any]:
        """获取所有数据"""
        self.logger.info(f"开始获取股票数据: {code}")

        result = {
            "code": code,
            "basic_info": self.get_basic_info(code),
            "price": self.get_latest_price(code),
            "financial_indicators": self.get_financial_indicators(code)
        }

        self.logger.info(f"数据获取完成: {code}")
        return result
