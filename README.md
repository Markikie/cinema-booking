# Cinema Booking

Cinema Booking is a full-stack movie seat booking application built for a concurrent booking assignment. It supports Google login, role-based admin access, showtime creation, seat locking, mock payment confirmation, real-time seat updates, and audit logs.

## Tech Stack

- Frontend: Vue 3, Vite, Pinia, Vue Router, Axios
- Backend: Go, Gin
- Database: MongoDB
- Locking and realtime events: Redis, Redis Pub/Sub, WebSocket
- Deployment: Docker Compose, nginx for frontend static hosting

## Features

- Google Sign-In login
- User and admin roles
- Admin can create movie showtimes and seed seat maps
- User can browse showtimes and select seats
- Seats are locked for a limited time before payment
- Selected seats can be unselected before checkout
- Other clients receive seat updates in real time through WebSocket
- Mock payment confirms bookings and marks seats as booked
- Expired locks release seats automatically
- Admin can view bookings and audit logs
- Seed command creates 10 sample movies with seats

## Project Structure

```text
cinema-booking/
  backend/              Go API, WebSocket, MongoDB/Redis integration
  frontend/             Vue frontend
  docker-compose.yml    MongoDB, Redis, backend, frontend services
```

## Environment

Create `backend/.env` from `backend/.env.example`.

Required values:

```env
PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=cinema_booking
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
GOOGLE_CLIENT_ID=your-google-oauth-client-id.apps.googleusercontent.com
JWT_SECRET=change-this-to-a-long-random-string
SEAT_LOCK_TTL_SECONDS=300
ADMIN_EMAILS=admin@example.com
ALLOWED_ORIGINS=http://localhost:5173
```

For frontend local development, create `frontend/.env`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
VITE_GOOGLE_CLIENT_ID=your-google-oauth-client-id.apps.googleusercontent.com
```

## Run With Docker

For Docker Compose, create a root `.env` file from `.env.example`:

```powershell
copy .env.example .env
```

Then set at least `JWT_SECRET`. Set `GOOGLE_CLIENT_ID` too if you want Google login to work from the Docker frontend.

From the repository root:

```powershell
docker-compose up --build
```

If no root `.env` exists, Docker Compose uses a development-only JWT secret so the backend can start. Do not use that default in production.

Services:

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`
- MongoDB: `localhost:27017`
- Redis: `localhost:6379`

## Run Locally

Start MongoDB and Redis first. Then run backend:

```powershell
cd backend
go run ./cmd/server
```

Run frontend:

```powershell
cd frontend
npm install
npm run dev -- --host 0.0.0.0
```

Open `http://localhost:5173`.

## Seed Data

Seed 10 sample movie showtimes and their seats:

```powershell
cd backend
go run ./cmd/seed
```

The seed command is idempotent. Running it again updates the same seeded showtimes and ensures their seats exist.

## Main API Flow

Authentication:

- `POST /api/auth/login` exchanges a Google ID token for an app JWT

Showtimes and seats:

- `GET /api/showtimes`
- `GET /api/showtimes/:showtime_id`
- `GET /api/showtimes/:showtime_id/seats`

Booking:

- `POST /api/bookings/select-seat`
- `POST /api/bookings/release-seat`
- `POST /api/bookings`
- `POST /api/bookings/confirm-payment`

Admin:

- `POST /api/admin/showtimes`
- `GET /api/admin/bookings`
- `GET /api/admin/audit-logs`

Realtime:

- `GET /ws?showtime_id=<id>&token=<jwt>`

## Concurrency Design

Seat selection uses Redis `SETNX` with a TTL to prevent multiple users from locking the same seat. The backend stores the locked state in MongoDB and publishes seat events through Redis Pub/Sub. WebSocket clients subscribed to the same showtime receive `SEAT_LOCKED`, `SEAT_RELEASED`, and `SEAT_BOOKED` updates.

When a lock expires, the expiry listener releases the seat, broadcasts the update, and marks pending bookings as timeout where applicable.

## Verification

Useful checks:

```powershell
cd backend
go test ./...
```

```powershell
cd frontend
npm run build
```

```powershell
docker-compose build frontend
```
