from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from app.models import models, schemas
from app.utils.database import get_db

router = APIRouter()

@router.get("/info", response_model=schemas.CarwashInfoResponse)
async def get_carwash_info(db: Session = Depends(get_db)):
    """
    Get information about the carwash (total boxes, available boxes, and their numbers)
    """
    # Get all boxes
    boxes = db.query(models.CarwashBox).all()
    
    # Get available boxes
    available_boxes = db.query(models.CarwashBox).filter(
        models.CarwashBox.status == models.BoxStatus.AVAILABLE.value
    ).all()
    
    # Extract available box numbers
    available_box_numbers = [box.box_number for box in available_boxes]
    
    return {
        "total_boxes": len(boxes),
        "available_boxes": len(available_boxes),
        "available_box_numbers": available_box_numbers
    }

@router.get("/boxes", response_model=List[schemas.CarwashBoxResponse])
async def get_all_boxes(db: Session = Depends(get_db)):
    """
    Get all carwash boxes
    """
    boxes = db.query(models.CarwashBox).all()
    return boxes

@router.get("/boxes/{box_id}", response_model=schemas.CarwashBoxResponse)
async def get_box(box_id: int, db: Session = Depends(get_db)):
    """
    Get a specific carwash box by ID
    """
    box = db.query(models.CarwashBox).filter(models.CarwashBox.id == box_id).first()
    if not box:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Box with ID {box_id} not found"
        )
    return box

@router.post("/boxes", response_model=schemas.CarwashBoxResponse, status_code=status.HTTP_201_CREATED)
async def create_box(box: schemas.CarwashBoxCreate, db: Session = Depends(get_db)):
    """
    Create a new carwash box (admin only)
    """
    # Check if box number already exists
    existing_box = db.query(models.CarwashBox).filter(
        models.CarwashBox.box_number == box.box_number
    ).first()
    
    if existing_box:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Box with number {box.box_number} already exists"
        )
    
    # Create new box
    new_box = models.CarwashBox(
        box_number=box.box_number,
        status=box.status
    )
    
    db.add(new_box)
    db.commit()
    db.refresh(new_box)
    
    return new_box

@router.put("/boxes/{box_id}", response_model=schemas.CarwashBoxResponse)
async def update_box(
    box_id: int, 
    box_update: schemas.CarwashBoxBase, 
    db: Session = Depends(get_db)
):
    """
    Update a carwash box status (admin only)
    """
    # Get the box
    box = db.query(models.CarwashBox).filter(models.CarwashBox.id == box_id).first()
    if not box:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Box with ID {box_id} not found"
        )
    
    # Update box
    box.status = box_update.status
    
    db.commit()
    db.refresh(box)
    
    return box
