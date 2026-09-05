"""Marketplace owns its own schema; never reuse core UUID users/orders tables."""
import os
import re
from sqlalchemy import create_engine, MetaData, text
from sqlalchemy.orm import declarative_base, sessionmaker
from app.core.config import DATABASE_URL
SCHEMA = os.getenv("MARKETPLACE_SCHEMA", "marketplace")
if not re.fullmatch(r"[a-z][a-z0-9_]{0,62}", SCHEMA) or SCHEMA in {"public", "pg_catalog", "information_schema"}:
    raise RuntimeError("MARKETPLACE_SCHEMA must be a dedicated, non-public schema")
engine = create_engine(DATABASE_URL, pool_pre_ping=True, connect_args={"connect_timeout": 10})
Base = declarative_base(metadata=MetaData(schema=SCHEMA))
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
def prepare_schema():
    with engine.begin() as connection:
        connection.execute(text(f'CREATE SCHEMA IF NOT EXISTS "{SCHEMA}"'))
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
