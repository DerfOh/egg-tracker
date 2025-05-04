version: '3.9'

services:
  backend:
    build:
      context: .
      dockerfile: ./docker/Dockerfile.backend
    volumes:
      - ./backend/data:/app/data
    ports:
      - "8080:8080"
    environment:
      - TZ=UTC
    depends_on:
      - duckdb
      - sqlite

  frontend:
    build:
      context: ./frontend
      dockerfile: ../docker/Dockerfile.frontend
    container_name: eggtracker-frontend
    ports:
      - "3000:3000"
    depends_on:
      - backend

  nginx:
    image: nginx:alpine
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - frontend
      - backend
    ports:
      - "80:80"

  duckdb:
    image: alpine
    container_name: duckdb-storage
    volumes:
      - ./backend/data:/data  # DuckDB persistence
    command: ["sh", "-c", "tail -f /dev/null"]

  sqlite:
    image: nouchka/sqlite3:latest
    container_name: sqlite-storage
    volumes:
      - ./backend/data:/data  # SQLite DB persistence
    command: ["tail", "-f", "/dev/null"]

volumes:
  db_data:

# Make sure to create the backend/data directory on your host before running docker-compose:
#   mkdir -p backend/data
# This ensures SQLite and DuckDB files are persisted and writable.