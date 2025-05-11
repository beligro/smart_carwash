from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession

from app.utils.config import get_settings

settings = get_settings()

# Convert the PostgreSQL URL to AsyncPG URL
SQLALCHEMY_DATABASE_URL = settings.database_url.replace("postgresql://", "postgresql+asyncpg://")

# Create async engine
engine = create_async_engine(
    SQLALCHEMY_DATABASE_URL,
    echo=True,
)

# Create async session
AsyncSessionLocal = sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

# Create base class for models
Base = declarative_base()

async def get_db():
    """Get database session."""
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()
