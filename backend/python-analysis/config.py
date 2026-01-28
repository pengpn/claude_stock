import os
from dotenv import load_dotenv

load_dotenv()

class Config:
    PORT = int(os.getenv('PYTHON_API_PORT', 5000))
    DEBUG = os.getenv('DEBUG', 'False').lower() == 'true'

config = Config()
