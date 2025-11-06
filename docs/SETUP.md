# 🚀 Couple Card Game - Setup Guide

This guide will help you set up and run the Couple Card Game application.

## Prerequisites

- **Go 1.22+** - [Install Go](https://golang.org/doc/install)
- **Node.js 18+** (for SASS compilation) - [Install Node.js](https://nodejs.org/)
- **Supabase Account** - [Sign up at Supabase](https://supabase.com/)
- **PostgreSQL** (managed by Supabase)

## Quick Start

### 1. Clone and Setup

```bash
# Navigate to project directory
cd /path/to/couple-card-game

# Install Go dependencies
go mod download

# Compile SASS to CSS
npx sass sass/main.scss static/css/main.css
```

### 2. Configure Supabase

1. Create a new project at [https://supabase.com](https://supabase.com/)
2. Go to **Settings** > **API** and copy:
   - **Project URL** (e.g., `https://your-project-id.supabase.co`)
   - **Anon/Public Key** (for client-side) or **Service Role Key** (for server-side)
3. Go to **SQL Editor** and run:
   - `sql/schema.sql` (database structure)
   - `sql/seed.sql` (sample data)

### 3. Environment Configuration

Create a `.env` file in the project root:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-anon-or-service-role-key

# Server Configuration
PORT=8080
ENV=development

# Session Configuration
SESSION_SECRET=change-this-to-a-random-secret-key-at-least-32-chars

# Admin Configuration
ADMIN_PASSWORD=admin123

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:8080,http://localhost:3000
```

⚠️ **Important**: Change `SESSION_SECRET` and `ADMIN_PASSWORD` in production!

### 4. Build and Run

```bash
# Build the application
go build -o server ./cmd/server

# Run the server
./server

# Or run directly without building
go run ./cmd/server/main.go
```

The server will start on `http://localhost:8080`

## 🎮 Using the Application

### For Players

1. **Visit Homepage**: `http://localhost:8080`
2. **Login Options**:
   - Click "Play as Guest" for anonymous play (4-hour session)
   - Click "Login" for OAuth (requires Supabase Auth setup)
3. **Create or Join Room**:
   - Create a new room and select categories
   - Or join an existing room with a room ID
4. **Play Game**:
   - Wait for a partner to join
   - Start the game and take turns answering questions

### For Administrators

1. **Access Admin Panel**: `http://localhost:8080/admin`
2. **Login**: Enter admin password (default: `admin123`)
3. **Manage**:
   - Users
   - Questions and Categories
   - Active rooms
   - View statistics

## 📁 Project Structure

```
couple-card-game/
├── cmd/
│   └── server/
│       └── main.go           # Application entrypoint
├── internal/
│   ├── handlers/             # HTTP request handlers
│   │   ├── base.go
│   │   ├── auth.go
│   │   ├── game.go
│   │   └── admin.go
│   ├── middleware/           # HTTP middleware
│   │   ├── auth.go
│   │   ├── admin.go
│   │   ├── session.go
│   │   └── ...
│   ├── models/               # Data models
│   │   ├── user.go
│   │   ├── room.go
│   │   ├── question.go
│   │   └── ...
│   └── services/             # Business logic
│       ├── supabase.go
│       ├── user_service.go
│       ├── game_service.go
│       └── ...
├── templates/                # HTML templates
│   ├── layout.html
│   ├── home.html
│   ├── auth/
│   ├── game/
│   └── admin/
├── static/                   # Static assets
│   ├── css/                  # Compiled CSS
│   ├── js/                   # JavaScript
│   └── i18n/                 # Translation files
├── sass/                     # SASS source files
│   ├── base/
│   ├── components/
│   └── pages/
├── sql/                      # SQL files
│   ├── schema.sql           # Database schema
│   ├── seed.sql             # Sample data
│   ├── migration_v2.sql     # Database migrations
│   └── fix_*.sql            # Database fixes
└── docker-compose.yml        # Docker setup
```

## 🔧 Development

### Compile SASS (Watch Mode)

```bash
npx sass --watch sass/main.scss:static/css/main.css
```

### Run with Live Reload

Using Air (Go live reload):

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with Air
air
```

### Database Migrations

To update the database schema:

1. Modify `sql/schema.sql`
2. Run in Supabase SQL Editor
3. Test with sample data from `sql/seed.sql`

## 🐳 Docker Setup

```bash
# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 🔐 Security Notes

### Production Checklist

- [ ] Change `SESSION_SECRET` to a strong random value
- [ ] Change `ADMIN_PASSWORD` to a strong password
- [ ] Use Supabase Service Role Key (not Anon Key)
- [ ] Enable Row Level Security (RLS) in Supabase
- [ ] Set proper `ALLOWED_ORIGINS` for CORS
- [ ] Use HTTPS in production
- [ ] Set `ENV=production`
- [ ] Review and restrict Supabase API permissions

### Session Security

- Sessions use secure HTTP-only cookies
- Anonymous sessions expire after 4 hours
- Session secret must be at least 32 characters
- Sessions are server-side managed

## 🌍 Internationalization

The app supports multiple languages (EN, FR, JA). Translation files are in `static/i18n/`:

- `en.json` - English
- `fr.json` - French  
- `ja.json` - Japanese

To add a new language:

1. Create `static/i18n/{lang}.json`
2. Copy structure from `en.json`
3. Translate all values
4. Add questions in the new language to the database

## 🧪 Testing

### Manual Testing

1. Start the server
2. Open `http://localhost:8080`
3. Test user flows:
   - Anonymous user creation
   - Room creation
   - Room joining
   - Game play
   - Admin panel

### Unit Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services/...
```

## 🐛 Troubleshooting

### Server won't start

- Check `.env` file exists and has correct values
- Verify Supabase credentials
- Ensure port 8080 is available
- Check logs for specific errors

### Database errors

- Verify Supabase project is active
- Check if `sql/schema.sql` was run successfully
- Ensure RLS policies are correctly configured
- Verify API key permissions

### CSS not loading

- Compile SASS: `npx sass sass/main.scss static/css/main.css`
- Check `static/css/main.css` exists
- Verify static file serving in browser DevTools

### Session issues

- Clear browser cookies
- Check `SESSION_SECRET` is set
- Verify session middleware is active

## 📚 API Endpoints

### Public Routes

- `GET /` - Home page
- `GET /health` - Health check
- `GET /auth/login` - Login page
- `POST /auth/logout` - Logout
- `POST /auth/anonymous` - Create anonymous user

### Game Routes (Authenticated)

- `GET/POST /game/create-room` - Create room
- `GET/POST /game/join-room` - Join room
- `GET /game/room/{id}` - Room lobby
- `GET /game/play/{id}` - Game play

### API Routes (HTMX)

- `POST /api/room/{id}/start` - Start game
- `POST /api/room/{id}/draw` - Draw question
- `POST /api/room/{id}/answer` - Submit answer
- `POST /api/room/{id}/finish` - Finish game

### Admin Routes

- `GET /admin` - Dashboard
- `GET /admin/users` - Manage users
- `GET /admin/questions` - Manage questions
- `GET /admin/categories` - Manage categories
- `GET /admin/rooms` - View rooms

## 🤝 Contributing

1. Follow existing code structure
2. Use KISS principle (Keep It Simple, Stupid)
3. Write clean, readable code
4. Test your changes
5. Update documentation

## 📝 License

This project is for educational and personal use.

## 🆘 Support

For issues or questions:
1. Check this SETUP.md
2. Review IMPLEMENTATION_STATUS.md
3. Check GitHub issues
4. Review Supabase documentation

---

**Ready to play! 💞**

