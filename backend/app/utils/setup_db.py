import asyncio
import logging
import subprocess
import sys
import time

from app.utils.init_db import init_db

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def run_migrations():
    """Run database migrations using Alembic."""
    logger.info("Running database migrations")
    
    try:
        # Run Alembic migrations
        subprocess.run(
            ["alembic", "upgrade", "head"],
            check=True,
            stdout=sys.stdout,
            stderr=sys.stderr,
        )
        logger.info("Database migrations completed successfully")
        return True
    except subprocess.CalledProcessError as e:
        logger.error(f"Error running database migrations: {e}")
        return False

async def setup_db():
    """Set up the database by running migrations and initializing data."""
    # Wait for the database to be ready
    logger.info("Waiting for database to be ready...")
    time.sleep(5)  # Simple delay to ensure database is up
    
    # Run migrations
    if run_migrations():
        # Initialize database with default data
        await init_db()
        logger.info("Database setup completed successfully")
    else:
        logger.error("Database setup failed")

if __name__ == "__main__":
    asyncio.run(setup_db())
