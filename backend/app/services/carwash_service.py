from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List, Dict, Any

from app.models.carwash_box import CarwashBox, BoxStatus
from app.utils.database import AsyncSessionLocal
from app.utils.schemas import CarwashInfo

async def get_all_boxes(db: AsyncSession = None) -> List[CarwashBox]:
    """Get all carwash boxes."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await get_all_boxes(db)
    
    result = await db.execute(select(CarwashBox).order_by(CarwashBox.id))
    return result.scalars().all()

async def get_box_by_id(box_id: int, db: AsyncSession = None) -> CarwashBox:
    """Get a carwash box by ID."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await get_box_by_id(box_id, db)
    
    result = await db.execute(select(CarwashBox).where(CarwashBox.id == box_id))
    return result.scalars().first()

async def create_box(name: str, status: BoxStatus = BoxStatus.AVAILABLE, db: AsyncSession = None) -> CarwashBox:
    """Create a new carwash box."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await create_box(name, status, db)
    
    box = CarwashBox(name=name, status=status)
    
    db.add(box)
    await db.commit()
    await db.refresh(box)
    
    return box

async def update_box_status(box_id: int, status: BoxStatus, db: AsyncSession = None) -> CarwashBox:
    """Update a carwash box status."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await update_box_status(box_id, status, db)
    
    box = await get_box_by_id(box_id, db)
    
    if box is None:
        return None
    
    box.status = status
    
    await db.commit()
    await db.refresh(box)
    
    return box

async def get_carwash_info(db: AsyncSession = None) -> CarwashInfo:
    """Get carwash information."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await get_carwash_info(db)
    
    boxes = await get_all_boxes(db)
    
    # Count boxes by status
    status_counts = {status: 0 for status in BoxStatus}
    for box in boxes:
        status_counts[box.status] += 1
    
    return CarwashInfo(
        boxes=boxes,
        total_boxes=len(boxes),
        available_boxes=status_counts[BoxStatus.AVAILABLE],
        occupied_boxes=status_counts[BoxStatus.OCCUPIED],
        maintenance_boxes=status_counts[BoxStatus.MAINTENANCE],
    )

async def initialize_default_boxes(db: AsyncSession = None) -> None:
    """Initialize default carwash boxes if none exist."""
    if db is None:
        async with AsyncSessionLocal() as db:
            await initialize_default_boxes(db)
            return
    
    # Check if any boxes exist
    result = await db.execute(select(func.count()).select_from(CarwashBox))
    count = result.scalar()
    
    if count == 0:
        # Create default boxes
        default_boxes = [
            {"name": "Бокс 1", "status": BoxStatus.AVAILABLE},
            {"name": "Бокс 2", "status": BoxStatus.AVAILABLE},
            {"name": "Бокс 3", "status": BoxStatus.AVAILABLE},
            {"name": "Бокс 4", "status": BoxStatus.MAINTENANCE},
        ]
        
        for box_data in default_boxes:
            await create_box(name=box_data["name"], status=box_data["status"], db=db)
