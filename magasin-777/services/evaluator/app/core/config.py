"""Application configuration."""

import os

DATABASE_URL: str = os.getenv(
    "DATABASE_URL",
    "",
)

SECRET_KEY: str = os.environ["SECRET_KEY"]
ALGORITHM: str = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES: int = 60 * 24  # 24 hours

UPLOAD_DIR: str = os.getenv("UPLOAD_DIR", "/app/static/uploads")

# Default admin credentials (created on first run)
DEFAULT_ADMIN_EMAIL: str = os.getenv("ADMIN_EMAIL", "admin@localmarket.com")
DEFAULT_ADMIN_PASSWORD: str = os.getenv("ADMIN_PASSWORD", "")
