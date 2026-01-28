from services.data_fetcher import StockDataFetcher
import json

if __name__ == '__main__':
    fetcher = StockDataFetcher()

    # 测试铜陵有色
    result = fetcher.fetch_all("000630")
    print(json.dumps(result, ensure_ascii=False, indent=2))
