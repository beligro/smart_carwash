from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.user import User
from app.utils.database import AsyncSessionLocal

async def get_user_by_telegram_id(telegram_id: str, db: AsyncSession = None) -> User:
    """Get a user by Telegram ID."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await get_user_by_telegram_id(telegram_id, db)
    
    result = await db.execute(select(User).where(User.telegram_id == telegram_id))
    return result.scalars().first()

async def create_user(
    telegram_id: str,
    username: str = None,
    first_name: str = None,
    last_name: str = None,
    is_admin: bool = False,
    db: AsyncSession = None,
) -> User:
    """Create a new user."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await create_user(
                telegram_id=telegram_id,
                username=username,
                first_name=first_name,
                last_name=last_name,
                is_admin=is_admin,
                db=db,
            )
    
    user = User(
        telegram_id=telegram_id,
        username=username,
        first_name=first_name,
        last_name=last_name,
        is_admin=is_admin,
    )
    
    db.add(user)
    await db.commit()
    await db.refresh(user)
    
    return user

async def create_user_if_not_exists(
    telegram_id: str,
    username: str = None,
    first_name: str = None,
    last_name: str = None,
    db: AsyncSession = None,
) -> User:
    """Create a user if it doesn't exist."""
    if db is None:
        async with AsyncSessionLocal() as db:
            return await create_user_if_not_exists(
                telegram_id=telegram_id,
                username=username,
                first_name=first_name,
                last_name=last_name,
                db=db,
            )
    
    user = await get_user_by_telegram_id(telegram_id, db)
    
    if user is None:
        user = await create_user(
            telegram_id=telegram_id,
            username=username,
            first_name=first_name,
            last_name=last_name,
            db=db,
        )
    
    return user
