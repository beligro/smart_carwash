from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from app.models import models, schemas
from app.utils.database import get_db

router = APIRouter()

@router.post("/", response_model=schemas.UserResponse, status_code=status.HTTP_201_CREATED)
async def create_user(user: schemas.UserCreate, db: Session = Depends(get_db)):
    """
    Create a new user (first time login)
    """
    # Check if user already exists
    existing_user = db.query(models.User).filter(
        models.User.telegram_id == user.telegram_id
    ).first()
    
    if existing_user:
        # Return existing user
        return existing_user
    
    # Create new user
    new_user = models.User(
        telegram_id=user.telegram_id,
        username=user.username,
        first_name=user.first_name,
        last_name=user.last_name,
        role=user.role
    )
    
    db.add(new_user)
    db.commit()
    db.refresh(new_user)
    
    return new_user

@router.get("/", response_model=List[schemas.UserResponse])
async def get_all_users(db: Session = Depends(get_db)):
    """
    Get all users (admin only)
    """
    users = db.query(models.User).all()
    return users

@router.get("/{user_id}", response_model=schemas.UserResponse)
async def get_user(user_id: int, db: Session = Depends(get_db)):
    """
    Get a specific user by ID
    """
    user = db.query(models.User).filter(models.User.id == user_id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User with ID {user_id} not found"
        )
    return user

@router.get("/telegram/{telegram_id}", response_model=schemas.UserResponse)
async def get_user_by_telegram_id(telegram_id: str, db: Session = Depends(get_db)):
    """
    Get a user by Telegram ID
    """
    user = db.query(models.User).filter(models.User.telegram_id == telegram_id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User with Telegram ID {telegram_id} not found"
        )
    return user

@router.put("/{user_id}", response_model=schemas.UserResponse)
async def update_user(
    user_id: int, 
    user_update: schemas.UserBase, 
    db: Session = Depends(get_db)
):
    """
    Update a user (admin only)
    """
    # Get the user
    user = db.query(models.User).filter(models.User.id == user_id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User with ID {user_id} not found"
        )
    
    # Update user fields
    for key, value in user_update.dict(exclude_unset=True).items():
        setattr(user, key, value)
    
    db.commit()
    db.refresh(user)
    
    return user

@router.put("/{user_id}/role", response_model=schemas.UserResponse)
async def update_user_role(
    user_id: int, 
    role: schemas.UserRole, 
    db: Session = Depends(get_db)
):
    """
    Update a user's role (admin only)
    """
    # Get the user
    user = db.query(models.User).filter(models.User.id == user_id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User with ID {user_id} not found"
        )
    
    # Update role
    user.role = role
    
    db.commit()
    db.refresh(user)
    
    return user
