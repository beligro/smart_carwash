import os
from fastapi import FastAPI, Depends
from fastapi.middleware.cors import CORSMiddleware

from app.routes import carwash, users
from app.services.telegram_bot import start_bot
from app.utils.config import get_settings

# Create FastAPI app
app = FastAPI(
    title="Smart Carwash API",
    description="API for Smart Carwash Telegram Mini App",
    version="1.0.0",
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, you should specify exact origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(carwash.router, prefix="/api", tags=["carwash"])
app.include_router(users.router, prefix="/api", tags=["users"])

@app.on_event("startup")
async def startup_event():
    """Start the Telegram bot when the application starts."""
    # Start the Telegram bot in a separate thread
    import asyncio
    from threading import Thread
    
    def run_bot():
        asyncio.run(start_bot())
    
    Thread(target=run_bot, daemon=True).start()

@app.get("/api/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "ok"}

if __name__ == "__main__":
    import uvicorn
    
    settings = get_settings()
    uvicorn.run(
        "app.main:app",
        host=settings.backend_host,
        port=settings.backend_port,
        reload=True,
    )
