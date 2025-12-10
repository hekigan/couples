# Makefile Commands Reference

All test database and testing operations are now accessible via Makefile commands for consistency.

---

## 📋 Quick Command Reference

### Most Common Commands

```bash
make help              # Show all available commands
make test-db-setup     # One-time test database setup
make test-full         # Run full test suite
make test-coverage-html # View coverage in browser
make dev               # Run server with hot-reload
make sass-watch        # Watch and compile SASS
make js-watch          # Watch and bundle JavaScript
```

---

## 🧪 Testing Commands

### Run Tests

| Command | Description | Usage |
|---------|-------------|-------|
| `make test` | Run short tests (unit tests only) | Daily development |
| `make test-short` | Alias for `make test` | Same as above |
| `make test-full` | Run full test suite (requires test DB) | Before commits |
| `make test-coverage` | Generate coverage report (terminal) | Check coverage % |
| `make test-coverage-html` | Generate and open coverage in browser | Detailed coverage view |

### Examples

```bash
# Quick check during development
make test

# Full test run before committing
make test-full

# Check coverage
make test-coverage

# View detailed coverage in browser
make test-coverage-html
```

---

## 🗄️ Test Database Commands

### Database Operations

| Command | Description | When to Use |
|---------|-------------|-------------|
| `make test-db-setup` | Complete one-time setup | First time only |
| `make test-db-start` | Start test database | Start of day |
| `make test-db-stop` | Stop test database | End of day |
| `make test-db-reset` | Reset database (clean slate) | Between test runs |
| `make test-db-status` | Show database status | Troubleshooting |
| `make test-db-studio` | Open Supabase Studio UI | View/edit data |

### Examples

```bash
# First time setup (includes CLI install, start, schema, seed)
make test-db-setup

# Daily workflow
make test-db-start      # Morning
make test-full          # Run tests
make test-db-stop       # Evening

# Clean slate for fresh test run
make test-db-reset

# View test data
make test-db-studio

# Check if database is running
make test-db-status
```

---

## 🏗️ Build & Run Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the Go binary (includes SASS + JS bundling) |
| `make run` | Run the server in production mode (builds everything) |
| `make dev` | Run in development mode with Air hot-reload |

---

## 🎨 Styling Commands

| Command | Description |
|---------|-------------|
| `make sass` | Compile SASS to CSS once |
| `make sass-watch` | Watch and auto-compile SASS on changes |

---

## 📦 JavaScript Bundling

| Command | Description | Usage |
|---------|-------------|-------|
| `make js-build` | Build production bundles (minified + source maps) | Production builds |
| `make js-build-dev` | Build development bundles (unminified) | Development mode |
| `make js-watch` | Watch and auto-bundle JavaScript on changes | Run in separate terminal |
| `make js-clean` | Remove all generated JavaScript bundles | Clean build artifacts |

### Bundle Information

- **Production bundles:**
  - `app.bundle.js` (61KB) + `app.bundle.js.map` (158KB)
  - `admin.bundle.js` (404B) + `admin.bundle.js.map` (4.5KB)
  - 82-94% size reduction vs development

- **Development bundles:**
  - `app.bundle.js` (358KB) with inline source maps
  - `admin.bundle.js` (6.7KB) with inline source maps

### Examples

```bash
# Production build
ENV=production make js-build

# Development build
ENV=development make js-build-dev

# Watch mode (auto-rebuild on file changes)
make js-watch

# Clean bundles
make js-clean
```

---

## 🐳 Docker Commands

| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker image |
| `make docker-run` | Start Docker container |
| `make docker-stop` | Stop Docker container |
| `make docker-logs` | View Docker logs |

---

## 🔧 Utility Commands

| Command | Description |
|---------|-------------|
| `make clean` | Clean build artifacts |
| `make deps` | Install dependencies (Go + Node) |
| `make fmt` | Format code |
| `make lint` | Lint code |
| `make db-setup` | Setup production database (manual) |
| `make setup` | Full setup (deps + sass + db) |

---

## 🚀 Common Workflows

### First Time Setup

```bash
# 1. Install all dependencies
make deps

# 2. Setup test database
make test-db-setup

# 3. Verify setup
make test-db-status

# 4. Run tests
make test
```

### Daily Development

```bash
# Morning - Start test database
make test-db-start

# Development - Run tests frequently
make test              # Quick check
make test-full         # Full suite

# View data if needed
make test-db-studio

# Evening - Stop test database
make test-db-stop
```

### Development with Hot-Reload (3 Terminals)

For full hot-reload experience during active development:

```bash
# Terminal 1: Go server with Air hot-reload
make dev

# Terminal 2: SASS watcher (auto-compile CSS)
make sass-watch

# Terminal 3: JavaScript watcher (auto-bundle JS)
make js-watch
```

**Note:** When you run `make dev`, a colored warning message will remind you about Terminals 2 and 3.

### Before Committing

```bash
# 1. Format code
make fmt

# 2. Run full test suite
make test-full

# 3. Check coverage
make test-coverage

# 4. Build to verify no errors
make build
```

### Troubleshooting

```bash
# Check test database status
make test-db-status

# Reset database
make test-db-reset

# View logs
supabase logs

# Clean and rebuild
make clean
make build
```

---

## 📊 Test Coverage Goals

| Service | Current | Target | Command |
|---------|---------|--------|---------|
| QuestionService | TBD | 80%+ | `make test-coverage` |
| AnswerService | TBD | 80%+ | `make test-coverage` |
| GameService | TBD | 90%+ | `make test-coverage` |
| FriendService | TBD | 85%+ | `make test-coverage` |
| UserService | TBD | 90%+ | `make test-coverage` |
| **Overall** | **0.8%** | **80%+** | `make test-coverage-html` |

**Note**: Current 0.8% is from short mode (unit tests only). Full integration tests await test database.

---

## 🎯 Command Cheat Sheet

### Quick Actions

```bash
make help                 # 📚 Show all commands
make test                 # 🧪 Quick test (unit only)
make test-full            # 🧪 Full test suite
make test-coverage-html   # 📊 Coverage in browser
make test-db-setup        # ⚙️  Setup test DB (once)
make test-db-start        # ▶️  Start test DB
make test-db-stop         # ⏹️  Stop test DB
make test-db-reset        # 🔄 Clean slate
make test-db-status       # ℹ️  Check DB status
make test-db-studio       # 🖥️  Open Studio UI
```

---

## 💡 Tips & Tricks

### Alias for Speed

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
alias mt='make test'
alias mtf='make test-full'
alias mtc='make test-coverage-html'
alias tdb='make test-db-start'
```

Then use:
```bash
mt      # Quick test
mtf     # Full test
mtc     # Coverage
tdb     # Start DB
```

### Watch Tests on Save

```bash
brew install watchexec
watchexec -e go -- make test-full
```

### Run Specific Test

```bash
# Use go test directly for specific tests
go test -v -run TestGetFriends ./internal/services/...
```

---

## 🆘 Troubleshooting

### Command Not Found

**Problem**: `make: supabase: No such file or directory`

**Solution**:
```bash
make test-db-setup  # Installs Supabase CLI automatically
```

### Database Connection Failed

**Problem**: Tests fail with connection error

**Solution**:
```bash
make test-db-status  # Check if running
make test-db-start   # Start if stopped
```

### Port Already in Use

**Problem**: Port 54321 already in use

**Solution**:
```bash
make test-db-stop
lsof -ti:54321 | xargs kill -9
make test-db-start
```

### Tests Still Skip

**Problem**: Tests skip with "database not configured"

**Solution**:
```bash
# Verify .env.test exists
cat .env.test

# Export variables
export $(cat .env.test | xargs)

# Verify DB running
make test-db-status
```

---

## 📚 Additional Resources

- **Quick Start**: [QUICK_START_TESTING.md](QUICK_START_TESTING.md)
- **Detailed Setup**: [TEST_DATABASE_SETUP.md](TEST_DATABASE_SETUP.md)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Implementation Complete**: [TEST_IMPLEMENTATION_COMPLETE.md](TEST_IMPLEMENTATION_COMPLETE.md)

---

## ✨ Summary

**All testing operations are now accessible via Makefile for consistency!**

**Most Used Commands:**
1. `make test-db-setup` - One time
2. `make test-db-start` - Daily
3. `make test-full` - Often
4. `make test-coverage-html` - Before commits
5. `make test-db-stop` - End of day

**Type `make help` to see all available commands anytime!**

---

**Last Updated**: November 2025
**Total Commands**: 20+ testing commands
**Coverage Target**: 80%+
