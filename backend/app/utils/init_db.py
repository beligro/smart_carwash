import logging
from sqlalchemy.orm import Session
from app.models import models, schemas
from app.utils.database import SessionLocal, create_tables

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def init_db():
    """
    Initialize the database with sample data
    """
    # Create tables
    create_tables()
    
    db = SessionLocal()
    try:
        # Check if we already have boxes
        existing_boxes = db.query(models.CarwashBox).count()
        if existing_boxes == 0:
            logger.info("Creating sample carwash boxes...")
            # Create 5 sample boxes
            for i in range(1, 6):
                box = models.CarwashBox(
                    box_number=i,
                    status=models.BoxStatus.AVAILABLE.value
                )
                db.add(box)
            
            # Create admin user
            admin_exists = db.query(models.User).filter(
                models.User.role == models.UserRole.ADMIN.value
            ).first()
            
            if not admin_exists:
                logger.info("Creating admin user...")
                admin = models.User(
                    telegram_id="admin_telegram_id",
                    username="admin",
                    first_name="Admin",
                    last_name="User",
                    role=models.UserRole.ADMIN.value
                )
                db.add(admin)
            
            db.commit()
            logger.info("Sample data created successfully")
        else:
            logger.info("Database already contains data, skipping initialization")
    except Exception as e:
        logger.error(f"Error initializing database: {e}")
    finally:
        db.close()

if __name__ == "__main__":
    logger.info("Initializing database...")
    init_db()
    logger.info("Database initialization completed")
