from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List

from app.utils.database import get_db
from app.utils.schemas import UserCreate, UserInDB, UserUpdate
from app.services.user_service import (
    get_user_by_telegram_id,
    create_user,
    create_user_if_not_exists,
)

router = APIRouter()

@router.post("/users", response_model=UserInDB)
async def create_new_user(user: UserCreate, db: AsyncSession = Depends(get_db)):
    """Create a new user."""
    # Check if user already exists
    existing_user = await get_user_by_telegram_id(user.telegram_id, db)
    
    if existing_user:
        raise HTTPException(status_code=400, detail="User already exists")
    
    return await create_user(
        telegram_id=user.telegram_id,
        username=user.username,
        first_name=user.first_name,
        last_name=user.last_name,
        db=db,
    )

@router.get("/users/{telegram_id}", response_model=UserInDB)
async def get_user(telegram_id: str, db: AsyncSession = Depends(get_db)):
    """Get a user by Telegram ID."""
    user = await get_user_by_telegram_id(telegram_id, db)
    
    if user is None:
        raise HTTPException(status_code=404, detail="User not found")
    
    return user

@router.post("/users/ensure", response_model=UserInDB)
async def ensure_user_exists(user: UserCreate, db: AsyncSession = Depends(get_db)):
    """Ensure a user exists, creating it if necessary."""
    return await create_user_if_not_exists(
        telegram_id=user.telegram_id,
        username=user.username,
        first_name=user.first_name,
        last_name=user.last_name,
        db=db,
    )
