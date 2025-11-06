# 🚀 Quick Start Guide

Get the Couple Card Game running in 5 minutes!

## Prerequisites

- Go 1.22 or higher
- Node.js (for SASS compilation)
- Supabase account

## 5-Minute Setup

### 1. Clone and Install

```bash
cd /path/to/couple-card-game
go mod download
npm install -g sass
```

### 2. Setup Supabase

1. Create a project at [supabase.com](https://supabase.com)
2. Go to **SQL Editor** and run:
   - `sql/schema.sql` (creates all tables)
   - `sql/seed.sql` (adds sample questions)
   - `sql/fix_rls_policies.sql` (fixes authentication)

### 3. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your Supabase credentials:

```env
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
SESSION_SECRET=your-random-secret-here
```

### 4. Build & Run

```bash
# Compile CSS
make sass

# Build server
make build

# Run server
./server
```

Or use the Makefile:

```bash
make dev  # Runs in development mode
```

### 5. Open Browser

```
http://localhost:8080
```

Click **"Play as Guest"** and start playing!

## Quick Commands

```bash
make help          # Show all commands
make sass          # Compile CSS
make sass-watch    # Watch CSS changes
make build         # Build binary
make run           # Run server
make dev           # Development mode
make docker-build  # Build Docker image
make docker-run    # Run with Docker
```

## Troubleshooting

### Database Connection Error

- Verify Supabase URL and key in `.env`
- Check if SQL scripts ran successfully
- Run `sql/fix_rls_policies.sql` if auth fails

### CSS Not Loading

```bash
make sass
```

### Port Already in Use

Change port in `.env`:

```env
PORT=8188
```

## Next Steps

- **Add OAuth**: See [docs/OAUTH_SETUP.md](docs/OAUTH_SETUP.md)
- **Configure Friends**: See [docs/FRIEND_SYSTEM.md](docs/FRIEND_SYSTEM.md)
- **Full Setup**: See [docs/SETUP.md](docs/SETUP.md)

---

**Ready to play!** 🎮💕

