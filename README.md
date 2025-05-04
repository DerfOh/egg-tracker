# Egg Tracker

A full-stack application for tracking eggs, inventory, and related data. Built with React (frontend), Go (backend), and DuckDB/SQLite for storage. Dockerized for easy development and deployment.

---

## Prerequisites
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

---

## Directory Structure

```
/egg-tracker
├── backend/         # Go backend source code
├── frontend/        # React frontend source code
├── docker/          # Dockerfiles and nginx config
├── data/            # Database files (persisted)
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Setup & Usage

### 1. Clone the repository
```
git clone <your-repo-url>
cd egg-tracker
```

### 2. Build and run with Docker Compose
```
docker-compose up --build
```
Or, if you have a Makefile:
```
make docker-up
```

### 3. Access the application
- Frontend: [http://localhost:3000](http://localhost:3000) or `http://<your-lan-ip>:3000`
- API: proxied via `/api` from the frontend (no direct access needed)

---

## Development Notes

- **Frontend**: React app, built and served by nginx in the frontend container.
- **Backend**: Go app, listens on port 8080 inside its container.
- **nginx**: Proxies `/api` requests from the frontend container to the backend container.
- **CORS**: In production, all requests go through nginx, so CORS is not an issue. For local development, the backend allows common origins (see `main.go`).
- **Database**: Data is persisted in the `data/` directory and mounted into the backend container.

---

## API Usage
- All API requests from the frontend should use relative paths (e.g., `/api/login`).
- Do not use hardcoded backend URLs in the frontend code.

---

## Production Deployment
- For HTTPS, use a reverse proxy (nginx, Caddy, Traefik) with SSL certificates in front of the frontend container.
- Restrict CORS origins in production to your real domain.
- Set cookies with `Secure` and `SameSite=None` for HTTPS.

---

## Troubleshooting
- If you get CORS or cookie issues, check your browser's dev tools and backend CORS config.
- If you change service names or ports, update the nginx config and CORS settings accordingly.

---

## Security Disclaimer

This project is intended for hobby or homebrew use. While basic security measures are in place, it is not recommended for production or public-facing deployments. For personal or small group use, the current login component is sufficient, but do not use as-is for sensitive applications. Always enforce HTTPS and consider additional security best practices if you expand the project.

## Inspiration

This project was inspired by the need to track the egg output of my birds. By recording daily egg counts, I can observe trends over time and gain a better understanding of their laying patterns and overall health. This helps in making informed decisions about their care and management.

---

## License
MIT
