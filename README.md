# 💞 Couple Card Game

A fun and engaging card game designed for couples to strengthen their relationship through meaningful conversations.

## ⚡ Quick Start

### For Development

```bash
# 1. Setup test database (one-time)
make test-db-setup

# 2. Run tests
make test

# 3. Development mode (3 terminals for full hot-reload)
make dev          # Terminal 1: Run server with Air
make sass-watch   # Terminal 2: Auto-compile SASS
make js-watch     # Terminal 3: Auto-bundle JavaScript

# OR: One-time build and run (production mode)
make build
make run
```

### For Testing

```bash
# See all available commands
make help

# Quick test commands
make test              # Run short tests
make test-full         # Run full test suite
make test-coverage     # View coverage
```

## 📚 Documentation

**All documentation is in the `docs/` folder:**

### Essential Guides
- **[docs/README.md](docs/README.md)** - Documentation index (start here!)
- **[docs/START_HERE.md](docs/START_HERE.md)** - Project overview
- **[docs/QUICKSTART.md](docs/QUICKSTART.md)** - 5-minute setup guide
- **[docs/SETUP.md](docs/SETUP.md)** - Detailed setup instructions
- **[docs/STATUS.md](docs/STATUS.md)** - Implementation status (100% complete)

### Testing Documentation
- **[docs/QUICK_START_TESTING.md](docs/QUICK_START_TESTING.md)** - 5-minute test setup
- **[docs/MAKEFILE_COMMANDS.md](docs/MAKEFILE_COMMANDS.md)** - All make commands
- **[docs/TESTING.md](docs/TESTING.md)** - Comprehensive testing guide
- **[docs/TEST_DATABASE_SETUP.md](docs/TEST_DATABASE_SETUP.md)** - Test DB setup

### Feature Guides
- **[docs/FRIEND_SYSTEM.md](docs/FRIEND_SYSTEM.md)** - Friend system documentation
- **[docs/OAUTH_SETUP.md](docs/OAUTH_SETUP.md)** - OAuth configuration
- **[docs/REALTIME_NOTIFICATIONS.md](docs/REALTIME_NOTIFICATIONS.md)** - SSE architecture

**👉 Start with [docs/README.md](docs/README.md) for the complete documentation index**

## ✨ Features

- 🎮 **Real-time gameplay** - Turn-based with SSE synchronization
- 🌍 **Multi-language support** - EN, FR, JA
- 👥 **Complete friend system** - Friend requests, invitations, search
- 🔐 **OAuth authentication** - Google, Facebook, GitHub
- 📱 **Mobile responsive** - Beautiful UI on all devices
- 🎨 **Professional polish** - Animations, toasts, loading states
- 🔄 **Reconnection handling** - Auto pause/resume
- 👤 **Admin panel** - User and content management
- ✅ **Comprehensive tests** - 58 test cases, 80%+ coverage target

## 🛠️ Tech Stack

- **Backend**: Go 1.22+
- **Database**: PostgreSQL (Supabase)
- **Frontend**: HTMX, SASS, JavaScript (bundled with esbuild)
- **Bundler**: esbuild (via Go API)
- **Real-time**: Server-Sent Events (SSE)
- **Testing**: Go test, Supabase CLI
- **Build**: Makefile, Docker

## 📊 Project Status

- **Status**: ✅ Production Ready (100% complete)
- **Code**: 8,500+ lines
- **Features**: 60+
- **API Endpoints**: 20+
- **Database Tables**: 12+
- **Documentation**: 15 files
- **Test Cases**: 58

See [docs/STATUS.md](docs/STATUS.md) for detailed status.

## 🚀 Quick Commands

```bash
make help              # See all available commands
make test-db-setup     # Setup test database (one-time)
make test              # Run tests
make test-coverage     # View coverage
make build             # Build application (includes JS bundling)
make run               # Run application (production mode)
make dev               # Run with hot-reload (requires 3 terminals)
make sass-watch        # Watch and compile SASS
make js-watch          # Watch and bundle JavaScript
make docker-build      # Build Docker image
```

## 📖 Learn More

- **Documentation Index**: [docs/README.md](docs/README.md)
- **Quick Start**: [docs/QUICKSTART.md](docs/QUICKSTART.md)
- **Test Setup**: [docs/QUICK_START_TESTING.md](docs/QUICK_START_TESTING.md)
- **All Commands**: [docs/MAKEFILE_COMMANDS.md](docs/MAKEFILE_COMMANDS.md)

## 📝 License

MIT



