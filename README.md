# teaparty

A random video + text chat app, similar to Omegle. Two strangers are matched together via WebRTC for live video and text chat. Either person can skip to be paired with someone new.

---

## How It Works

1. A user opens the app and clicks **Start**. Their browser requests camera access and connects to the signalling server via WebSocket.
2. The Go server puts them in a waiting queue. When two users are queued, it pairs them into a **Room** and tells each one their role (`offerer` or `answerer`).
3. The server relays WebRTC signalling messages (offer, answer, ICE candidates) between the two clients until a peer-to-peer connection is established.
4. Video streams and text messages flow **directly peer-to-peer** over WebRTC, so the server is no longer in the media path.
5. Either user can hit **Skip** to drop the connection and be re-queued for a new match.

---

## Stack

| Layer    | Tech                                                                 |
|----------|----------------------------------------------------------------------|
| Server   | Go, [chi](https://github.com/go-chi/chi), [coder/websocket](https://github.com/coder/websocket), [godotenv](https://github.com/joho/godotenv) |
| Client   | Next.js 16, React 19, TypeScript, Tailwind CSS v4, Biome            |
| Media    | WebRTC (browser-native), TURN via [Metered](https://www.metered.ca) |

---

## Project Structure

```
teaparty/
├── server/
│   ├── cmd/server/main.go           # Entry point, listens on provided address
│   └── internal/
│       ├── app/app.go               # Queue logic, room management, TURN fetching
│       ├── api/websocket/handler.go # WebSocket handler + signalling relay
│       ├── room/room.go             # Room and Client types, write pump
│       └── routes/routes.go         # /health and /ws routes
└── client/
    ├── app/                         # Next.js app router
    └── components/
        ├── Chat.tsx                 # Core state: WebSocket, RTCPeerConnection, streams
        ├── VideoChat.tsx            # Local + remote video display
        └── TextChat.tsx             # Text messages over RTCDataChannel
```

---

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.25+
- [Node.js](https://nodejs.org/) 18+ and npm
- A [Metered](https://www.metered.ca) account for TURN servers (optional, falls back to Google STUN)

### Server

```bash
cd server

# copy and fill in environment variables
cp .env.example .env
```

Edit `.env`:

```env
METERED_DOMAIN=your-app.metered.live
METERED_SECRET_KEY=your_secret_key
TURN_CREDENTIALS_EXPIRY_SECONDS=86400
```

```bash
go run ./cmd/server 0.0.0.0:8080
```

The server exposes:
- `GET /health`: health check
- `GET /ws`: WebSocket signalling endpoint

### Client

```bash
cd client

# copy and fill in environment variables
cp .env.example .env.local
```

Edit `.env.local`:

```env
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in two browser tabs (or two devices on the same network) to test a match.

---

## Environment Variables

| Variable | Location | Description |
|---|---|---|
| `METERED_DOMAIN` | server | Your Metered app domain |
| `METERED_SECRET_KEY` | server | Metered secret key for TURN credentials |
| `TURN_CREDENTIALS_EXPIRY_SECONDS` | server | How long TURN credentials stay valid |
| `NEXT_PUBLIC_WS_URL` | client | WebSocket URL of the signalling server |

> If TURN credentials can't be fetched (e.g. no `.env` set), the client falls back to Google's public STUN server (`stun:stun.l.google.com:19302`). This works on most local networks but may fail across certain NATs.

---

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes
4. Open a pull request
