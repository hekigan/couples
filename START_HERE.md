# 🎮 Couple Card Game - START HERE

**Version**: 1.0.0  
**Status**: ✅ Production-Ready  
**Last Updated**: November 6, 2025

---

## 👋 Welcome!

This is your **complete, production-ready** Couple Card Game application. Everything is implemented, documented, and ready to run.

---

## ⚡ Quick Start (5 minutes)

```bash
# 1. Setup Supabase (https://supabase.com)
# 2. Run sql/schema.sql and sql/seed.sql in SQL Editor
# 3. Create .env file with your credentials

# 4. Compile CSS
npx sass sass/main.scss static/css/main.css

# 5. Run server
go run ./cmd/server/main.go

# 6. Open browser
open http://localhost:8080
```

---

## 📚 Documentation

All documentation is located in the **`docs/`** folder:

### Essential Guides
| Document | Purpose | Read When |
|----------|---------|-----------|
| **[docs/QUICKSTART.md](docs/QUICKSTART.md)** | Get running in 5 minutes | Starting now |
| **[docs/SETUP.md](docs/SETUP.md)** | Comprehensive setup guide | Setting up for development |
| **[docs/PROJECT_STATUS.md](docs/PROJECT_STATUS.md)** | Complete implementation status | Checking what's done |
| **[docs/README.md](docs/README.md)** | Project overview & features | Understanding the project |
| **[docs/OAUTH_SETUP.md](docs/OAUTH_SETUP.md)** | OAuth configuration guide | Setting up social login |
| **[docs/plan.md](docs/plan.md)** | Original specification | Understanding requirements |

---

## ✅ What's Included

### Core Features
- ✅ **User Authentication** - OAuth + Anonymous users
- ✅ **Real-time Gameplay** - Server-Sent Events (SSE)
- ✅ **Multi-language** - EN, FR, ES support
- ✅ **Friend System** - Send/accept friend requests
- ✅ **Room Management** - Create, join, manage rooms
- ✅ **Question System** - 50+ questions across 5 categories
- ✅ **Admin Panel** - Full content management
- ✅ **CSV Import/Export** - Bulk question management
- ✅ **Translation Management** - Admin UI for translations
- ✅ **Room Join Requests** - Request-based room joining

### Technical Features
- ✅ **Go Backend** - Clean architecture with services
- ✅ **PostgreSQL/Supabase** - Managed database
- ✅ **HTMX Frontend** - Dynamic without heavy JavaScript
- ✅ **SASS Styling** - Modern, responsive design
- ✅ **Row Level Security** - Database-level permissions
- ✅ **Docker Support** - Containerized deployment
- ✅ **Comprehensive Tests** - Unit and integration tests

---

## 🗂️ Project Structure

```
couple-card-game/
├── cmd/server/              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic
│   ├── models/              # Data models
│   └── middleware/          # HTTP middleware
├── templates/               # HTML templates
│   ├── game/                # Game screens
│   ├── admin/               # Admin panels
│   ├── auth/                # Auth pages
│   ├── friends/             # Friend management
│   └── components/          # Reusable components
├── static/                  # Static assets
│   ├── css/                 # Compiled CSS
│   ├── js/                  # JavaScript files
│   └── i18n/                # Translation files
├── sass/                    # SASS source files
├── sql/                     # Database scripts
│   ├── schema.sql           # Database schema
│   ├── seed.sql             # Sample data
│   ├── migration_v2.sql     # Database migrations
│   └── fix_*.sql            # Database fixes
└── docs/                    # Documentation
```

---

## 🚀 Deployment

### Prerequisites
- Go 1.22+
- PostgreSQL (or Supabase account)
- Node.js (for SASS compilation)

### Production Checklist
- [ ] Configure `.env` with production credentials
- [ ] Run database migrations
- [ ] Compile CSS: `npx sass sass/main.scss static/css/main.css`
- [ ] Build binary: `go build ./cmd/server/main.go`
- [ ] Configure OAuth providers (Google, Facebook)
- [ ] Set up SSL/TLS certificates
- [ ] Configure firewall and security

---

## 🎯 Key Features Explained

### Real-time Gameplay
Players see updates instantly without page refresh:
- Player joins/leaves
- Questions drawn
- Answers submitted
- Turn changes
- Game completed

### Multi-language Support
Complete i18n system:
- UI translated in EN, FR, ES
- Questions in multiple languages
- Admin UI for translation management
- Easy to add new languages

### Admin Panel
Comprehensive content management:
- **Users**: View, manage users
- **Questions**: Add, edit, delete questions
- **Categories**: Manage question categories
- **Rooms**: Monitor active game rooms
- **Translations**: Edit all UI text
- **CSV Import/Export**: Bulk operations

### Friend System
Social features for couples:
- Send friend requests
- Accept/reject requests
- Invite friends to rooms
- Track game history with friends

### Room Management
Flexible game rooms:
- Public or private rooms
- Room join requests (owner approval)
- Room deletion with cleanup
- 2-room limit per user
- Real-time player updates

---

## 📖 Learning Resources

### For Developers
1. Start with **[docs/SETUP.md](docs/SETUP.md)** for environment setup
2. Review **[docs/PROJECT_STATUS.md](docs/PROJECT_STATUS.md)** for architecture
3. Read code in `internal/` for implementation details
4. Check tests in `*_test.go` files

### For Administrators
1. Access admin panel at `/admin`
2. Use CSV import/export for bulk question updates
3. Manage translations through the UI
4. Monitor active rooms and users

### For Content Creators
1. Export questions to CSV
2. Edit in spreadsheet (Excel, Google Sheets)
3. Import updated CSV
4. Changes are live immediately

---

## 🆘 Need Help?

### Common Issues
- **Database connection failed**: Check `.env` credentials
- **CSS not loading**: Run `npx sass sass/main.scss static/css/main.css`
- **Anonymous users fail**: Run `sql/fix_rls_policies.sql`
- **OAuth not working**: Check `docs/OAUTH_SETUP.md`

### Getting Support
- Check **[docs/SETUP.md](docs/SETUP.md)** for troubleshooting
- Review **[docs/PROJECT_STATUS.md](docs/PROJECT_STATUS.md)** for known issues
- Consult inline code documentation
- Check Supabase dashboard for database issues

---

## 🎉 You're Ready!

The application is **100% complete and production-ready**. Everything works, is documented, and tested.

**Next Steps:**
1. Configure your environment (`.env`)
2. Run the quick start commands above
3. Open http://localhost:8080
4. Start playing! 🎮

---

**Happy Gaming!** 💕
