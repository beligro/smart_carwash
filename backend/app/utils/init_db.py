import asyncio
import logging
from sqlalchemy.ext.asyncio import AsyncSession

from app.utils.database import AsyncSessionLocal
from app.services.carwash_service import initialize_default_boxes

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def init_db():
    """Initialize the database with default data."""
    logger.info("Creating initial data")
    
    async with AsyncSessionLocal() as db:
        # Initialize default carwash boxes
        await initialize_default_boxes(db)
    
    logger.info("Initial data created")

if __name__ == "__main__":
    asyncio.run(init_db())
