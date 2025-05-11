from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List

from app.utils.database import get_db
from app.utils.schemas import CarwashInfo, CarwashBoxInDB, CarwashBoxUpdate
from app.services.carwash_service import (
    get_carwash_info,
    get_box_by_id,
    update_box_status,
)

router = APIRouter()

@router.get("/carwash/info", response_model=CarwashInfo)
async def get_carwash_information(db: AsyncSession = Depends(get_db)):
    """Get information about the carwash."""
    return await get_carwash_info(db)

@router.get("/carwash/boxes/{box_id}", response_model=CarwashBoxInDB)
async def get_carwash_box(box_id: int, db: AsyncSession = Depends(get_db)):
    """Get a specific carwash box by ID."""
    box = await get_box_by_id(box_id, db)
    
    if box is None:
        raise HTTPException(status_code=404, detail="Box not found")
    
    return box

@router.patch("/carwash/boxes/{box_id}", response_model=CarwashBoxInDB)
async def update_carwash_box_status(
    box_id: int, box_update: CarwashBoxUpdate, db: AsyncSession = Depends(get_db)
):
    """Update a carwash box status."""
    if box_update.status is None:
        raise HTTPException(status_code=400, detail="Status is required")
    
    box = await update_box_status(box_id, box_update.status, db)
    
    if box is None:
        raise HTTPException(status_code=404, detail="Box not found")
    
    return box
