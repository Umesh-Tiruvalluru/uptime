# Better Uptime frontend

The Compose stack builds and serves this application at `http://localhost:3000`.

For local development outside Docker, copy `.env.example` to `.env.local`, then run:

```bash
npm install
npm run dev
```

Set `NEXT_PUBLIC_API_URL` to the API's browser-accessible address (normally `http://localhost:8080`).
