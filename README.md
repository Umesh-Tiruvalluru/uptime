# Uptime App

## Run the full stack

1. Start Docker Desktop.
2. Set a strong `JWT_SECRET` in your existing `.env`. If you do not have one,
   create it from `.env.example`.
3. Run:

   ```powershell
   docker compose up --build
   ```

Compose starts PostgreSQL, NATS, runs Goose migrations once, then starts the API,
worker, and frontend.

| Service | Address |
| --- | --- |
| Frontend | http://localhost:3000 |
| Public status | http://localhost:3000/status |
| API | http://localhost:8080 |
| NATS | nats://localhost:4222 |

Stop the stack with `docker compose down`. The PostgreSQL volume is retained. To
remove the local database too, use `docker compose down -v`.
