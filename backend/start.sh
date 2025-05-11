#!/bin/bash
set -e

# Run database setup
python -m app.utils.setup_db

# Start the application
exec uvicorn app.main:app --host 0.0.0.0 --port 8000
