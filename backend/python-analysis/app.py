from flask import Flask, jsonify
from flask_cors import CORS
from config import config
from utils.logger import logger

app = Flask(__name__)
CORS(app)

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ok", "service": "python-analysis"})

if __name__ == '__main__':
    logger.info(f"Starting Python Analysis Service on port {config.PORT}")
    app.run(host='0.0.0.0', port=config.PORT, debug=config.DEBUG)
