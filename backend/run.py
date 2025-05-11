import os
import argparse
import uvicorn
import logging
from dotenv import load_dotenv
from app.utils.init_db import init_db

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)
logger = logging.getLogger(__name__)

def run_api():
    """
    Run the FastAPI application
    """
    host = os.getenv("API_HOST", "0.0.0.0")
    port = int(os.getenv("API_PORT", "8000"))
    
    logger.info(f"Starting API server on {host}:{port}")
    uvicorn.run("app.main:app", host=host, port=port, reload=True)

def run_bot():
    """
    Run the Telegram bot
    """
    logger.info("Starting Telegram bot")
    from app.bot import main
    main()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Smart Carwash backend services")
    parser.add_argument(
        "--service", 
        choices=["api", "bot", "init-db", "all"], 
        default="api",
        help="Service to run (api, bot, init-db, or all)"
    )
    
    args = parser.parse_args()
    
    if args.service == "init-db" or args.service == "all":
        logger.info("Initializing database...")
        init_db()
    
    if args.service == "api" or args.service == "all":
        run_api()
    
    if args.service == "bot" or args.service == "all":
        run_bot()
