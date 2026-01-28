from flask import Flask, jsonify, request
from flask_cors import CORS
from config import config
from utils.logger import logger
from services.data_fetcher import StockDataFetcher
from services.financial_analyzer import FinancialAnalyzer

app = Flask(__name__)
CORS(app)

# 初始化服务
data_fetcher = StockDataFetcher()
financial_analyzer = FinancialAnalyzer()

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ok", "service": "python-analysis"})

@app.route('/analyze', methods=['POST'])
def analyze():
    try:
        data = request.get_json()
        code = data.get('code')

        if not code:
            return jsonify({"error": "缺少股票代码"}), 400

        # 获取数据
        stock_data = data_fetcher.fetch_all(code)

        if "error" in stock_data.get("basic_info", {}):
            return jsonify({"error": "股票代码不存在或数据获取失败"}), 404

        # 财务分析
        analysis = financial_analyzer.analyze(stock_data)

        # 合并结果
        result = {
            "code": code,
            "name": stock_data["basic_info"].get("name", ""),
            "basic_info": stock_data["basic_info"],
            "price": stock_data["price"],
            **analysis
        }

        return jsonify(result)

    except Exception as e:
        logger.error(f"分析失败: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    logger.info(f"Starting Python Analysis Service on port {config.PORT}")
    app.run(host='0.0.0.0', port=config.PORT, debug=config.DEBUG)
