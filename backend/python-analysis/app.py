from flask import Flask, jsonify, request
from flask_cors import CORS
from config import config
from utils.logger import logger
from services.data_fetcher import StockDataFetcher
from services.financial_analyzer import FinancialAnalyzer
import akshare as ak

app = Flask(__name__)
CORS(app)

# 初始化服务
data_fetcher = StockDataFetcher()
financial_analyzer = FinancialAnalyzer()

# 缓存股票列表
_stock_list_cache = None

def get_stock_code_by_name(name_or_code):
    """
    根据股票名称或代码获取股票代码
    如果输入是6位数字，直接返回
    否则在股票列表中查找匹配的名称
    """
    global _stock_list_cache

    # 如果是6位数字，直接返回
    if name_or_code.isdigit() and len(name_or_code) == 6:
        return name_or_code

    try:
        # 获取股票列表（使用缓存）
        if _stock_list_cache is None:
            logger.info("加载股票列表...")
            _stock_list_cache = ak.stock_info_a_code_name()

        # 在股票列表中查找名称匹配
        matched = _stock_list_cache[_stock_list_cache['name'].str.contains(name_or_code, na=False)]

        if len(matched) > 0:
            code = matched.iloc[0]['code']
            stock_name = matched.iloc[0]['name']
            logger.info(f"名称匹配成功: '{name_or_code}' -> {code} ({stock_name})")
            return code
        else:
            logger.warning(f"未找到匹配的股票: '{name_or_code}'")
            return None

    except Exception as e:
        logger.error(f"股票名称查询失败: {e}")
        return None

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ok", "service": "python-analysis"})

@app.route('/analyze', methods=['POST'])
def analyze():
    try:
        data = request.get_json()
        input_value = data.get('code')

        if not input_value:
            return jsonify({"error": "缺少股票代码或名称"}), 400

        # 将名称转换为代码
        code = get_stock_code_by_name(input_value)

        if not code:
            return jsonify({"error": f"未找到股票: {input_value}"}), 404

        # 获取数据
        stock_data = data_fetcher.fetch_all(code)

        if "error" in stock_data.get("basic_info", {}):
            return jsonify({"error": "股票代码不存在或数据获取失败"}), 404

        # 财务分析
        analysis = financial_analyzer.analyze(stock_data)

        # 合并结果，确保与 Go 端 PythonAnalysisResponse 结构匹配
        result = {
            "code": code,
            "name": stock_data["basic_info"].get("name", ""),
            "basic_info": stock_data["basic_info"],
            "price": stock_data["price"],
            **analysis,  # financial_metrics + risks
        }

        logger.info(f"返回分析结果: input={input_value}, code={code}, name={result['name']}, "
                     f"price={result['price'].get('latest_price')}, "
                     f"pe={result['basic_info'].get('pe_ttm')}, "
                     f"pb={result['basic_info'].get('pb')}")

        return jsonify(result)

    except Exception as e:
        logger.error(f"分析失败: {e}", exc_info=True)
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    logger.info(f"Starting Python Analysis Service on port {config.PORT}")
    app.run(host='0.0.0.0', port=config.PORT, debug=config.DEBUG)
