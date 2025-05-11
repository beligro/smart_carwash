from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime
from enum import Enum

class UserRole(str, Enum):
    ADMIN = "admin"
    USER = "user"

class BoxStatus(str, Enum):
    AVAILABLE = "available"
    OCCUPIED = "occupied"
    MAINTENANCE = "maintenance"

class UserBase(BaseModel):
    telegram_id: str
    username: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None

class UserCreate(UserBase):
    role: UserRole = UserRole.USER

class UserResponse(UserBase):
    id: int
    role: UserRole
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class CarwashBoxBase(BaseModel):
    box_number: int
    status: BoxStatus = BoxStatus.AVAILABLE

class CarwashBoxCreate(CarwashBoxBase):
    pass

class CarwashBoxResponse(CarwashBoxBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class CarwashInfoResponse(BaseModel):
    total_boxes: int
    available_boxes: int
    available_box_numbers: List[int]
