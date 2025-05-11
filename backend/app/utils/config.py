import os
from functools import lru_cache
from pydantic_settings import BaseSettings
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

class Settings(BaseSettings):
    """Application settings."""
    
    # PostgreSQL settings
    postgres_user: str = os.getenv("POSTGRES_USER", "postgres")
    postgres_password: str = os.getenv("POSTGRES_PASSWORD", "postgres_password")
    postgres_db: str = os.getenv("POSTGRES_DB", "carwash_db")
    postgres_host: str = os.getenv("POSTGRES_HOST", "postgres")
    postgres_port: int = int(os.getenv("POSTGRES_PORT", "5432"))
    
    # Database URL
    database_url: str = os.getenv(
        "DATABASE_URL", 
        f"postgresql://{postgres_user}:{postgres_password}@{postgres_host}:{postgres_port}/{postgres_db}"
    )
    
    # Backend settings
    backend_host: str = os.getenv("BACKEND_HOST", "0.0.0.0")
    backend_port: int = int(os.getenv("BACKEND_PORT", "8000"))
    
    # Telegram settings
    telegram_bot_token: str = os.getenv("TELEGRAM_BOT_TOKEN", "")
    telegram_bot_name: str = os.getenv("TELEGRAM_BOT_NAME", "")
    server_ip: str = os.getenv("SERVER_IP", "")
    
    class Config:
        env_file = ".env"
        case_sensitive = False

@lru_cache()
def get_settings() -> Settings:
    """Get application settings."""
    return Settings()
