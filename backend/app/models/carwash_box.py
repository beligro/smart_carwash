from sqlalchemy import Column, Integer, String, Enum, DateTime
from sqlalchemy.sql import func
import enum

from app.utils.database import Base

class BoxStatus(str, enum.Enum):
    """Enum for carwash box status."""
    
    AVAILABLE = "available"
    OCCUPIED = "occupied"
    MAINTENANCE = "maintenance"

class CarwashBox(Base):
    """Carwash box model."""
    
    __tablename__ = "carwash_boxes"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    status = Column(Enum(BoxStatus), default=BoxStatus.AVAILABLE, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    def __repr__(self):
        return f"<CarwashBox(id={self.id}, name={self.name}, status={self.status})>"
