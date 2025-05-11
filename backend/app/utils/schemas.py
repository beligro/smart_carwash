from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime
from enum import Enum

# User schemas
class UserBase(BaseModel):
    telegram_id: str
    username: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None

class UserCreate(UserBase):
    pass

class UserUpdate(BaseModel):
    username: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    is_admin: Optional[bool] = None

class UserInDB(UserBase):
    id: int
    is_admin: bool
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

# Carwash box schemas
class BoxStatus(str, Enum):
    AVAILABLE = "available"
    OCCUPIED = "occupied"
    MAINTENANCE = "maintenance"

class CarwashBoxBase(BaseModel):
    name: str
    status: BoxStatus = BoxStatus.AVAILABLE

class CarwashBoxCreate(CarwashBoxBase):
    pass

class CarwashBoxUpdate(BaseModel):
    name: Optional[str] = None
    status: Optional[BoxStatus] = None

class CarwashBoxInDB(CarwashBoxBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

# Carwash info schema
class CarwashInfo(BaseModel):
    boxes: List[CarwashBoxInDB]
    total_boxes: int
    available_boxes: int
    occupied_boxes: int
    maintenance_boxes: int
