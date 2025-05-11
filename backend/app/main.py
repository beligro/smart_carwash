from fastapi import FastAPI, Depends
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
from app.routes import carwash, users
from app.utils.database import create_tables

app = FastAPI(
    title="Smart Carwash API",
    description="API for Smart Carwash Telegram Mini App",
    version="1.0.0"
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, replace with specific origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(carwash.router, prefix="/api/carwash", tags=["carwash"])
app.include_router(users.router, prefix="/api/users", tags=["users"])

@app.on_event("startup")
async def startup_event():
    # Create database tables on startup
    create_tables()

@app.get("/")
async def root():
    return {"message": "Smart Carwash API is running"}

if __name__ == "__main__":
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)
