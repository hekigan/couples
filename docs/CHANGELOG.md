# 📅 Changelog - Couples Card Game

All notable changes and milestones for this project.

---

## [1.0.0] - November 2025 - PRODUCTION RELEASE ✅

### Major Milestones

**Phase 1: Core Game Implementation**
- ✅ Complete database schema with Supabase integration
- ✅ Question/Answer services with history tracking
- ✅ Turn-based gameplay logic
- ✅ Category selection with real-time sync
- ✅ Game API endpoints (7 total)
- ✅ Play interface with SSE
- ✅ Results page with statistics

**Phase 2: Friend System**
- ✅ Complete friend service (bidirectional queries)
- ✅ Friend request workflow
- ✅ Search users by username
- ✅ Room invitation system
- ✅ Notification integration

**Phase 3: Security & Admin**
- ✅ Admin password authentication
- ✅ User management with cascade delete
- ✅ Anonymous user cleanup (3 strategies)
- ✅ Session-based access control
- ✅ Documentation sanitization

**Phase 4: Reconnection & Polish**
- ✅ Game pause/resume on disconnect
- ✅ Timeout handling
- ✅ Global animation system
- ✅ Toast notification library
- ✅ Loading states throughout
- ✅ Mobile responsive design
- ✅ 15+ smooth animations

---

## [0.9.0] - November 2025 - Pre-Release

### Added
- OAuth integration (Google, Facebook, GitHub)
- i18n support (EN, FR, JA)
- SASS styling system
- Docker deployment configuration
- Comprehensive documentation (12+ guides)

### Fixed
- Session management security
- Cascade delete for user removal
- Real-time category synchronization
- Mobile layout issues

---

## [0.5.0] - November 2025 - Alpha

### Added
- Initial project structure
- Database schema with RLS
- Basic user authentication
- Room creation and joining
- Question/Answer models
- Admin panel foundation

### Infrastructure
- Go 1.22 setup
- Supabase integration
- HTMX for frontend interactivity
- Environment configuration

---

## Key Features by Release

### v1.0.0 (Current)
- **Gameplay**: Turn-based multiplayer with real-time updates
- **Social**: Complete friend system with invitations
- **Admin**: Full content management panel
- **Security**: Password-protected admin, cascade deletes
- **UX**: Professional animations, toast notifications, mobile-responsive
- **Reconnection**: Auto-pause/resume on disconnect

### v0.9.0
- **Auth**: OAuth providers + anonymous users
- **i18n**: Multi-language support
- **Deployment**: Docker-ready with documentation

### v0.5.0
- **Foundation**: Core architecture and database
- **Basic Gameplay**: Question drawing and answering
- **Admin**: Initial panel for content management

---

## Technical Achievements

**Backend**
- 8,500+ lines of Go code
- 12+ database tables
- 20+ API endpoints
- 8+ service layers
- Full CRUD operations

**Frontend**
- HTMX-based interactivity
- Server-Sent Events (SSE) for real-time
- SASS for maintainable CSS
- 15+ animations
- Toast notification system
- Mobile-first responsive design

**DevOps**
- Docker containerization
- Environment-based configuration
- Graceful shutdown handling
- Static asset optimization

---

## Breaking Changes

### v1.0.0
- Room model extended with `paused_at` and `disconnected_user` fields
- Friend model now includes `status` field (migration required)
- Answer model requires `action_type` field (answered/passed)

### v0.9.0
- Session cookie configuration changed (requires re-authentication)
- OAuth callback URLs updated

---

## Performance Improvements

### v1.0.0
- Optimized real-time category sync (reduced payload size)
- Added connection pooling for database
- Implemented lazy loading for templates
- Reduced JavaScript bundle size

### v0.9.0
- Added database indexes for common queries
- Implemented query result caching
- Optimized friend list queries (bidirectional)

---

## Security Updates

### v1.0.0
- ✅ Admin password authentication
- ✅ Documentation credentials sanitized
- ✅ Session security hardened
- ✅ CSRF protection enabled
- ✅ XSS prevention in templates

### v0.9.0
- OAuth token validation
- Secure cookie configuration
- SQL injection prevention
- Input sanitization

---

## Documentation Updates

### v1.0.0 (November 2025)
- **Cleanup**: Removed 6 redundant restoration logs
- **Consolidated**: Merged CURRENT_STATUS + PROJECT_STATUS → STATUS.md
- **Sanitized**: Removed hardcoded credentials from SETUP.md
- **Added**: CHANGELOG.md (this file)
- **Updated**: README.md with new navigation

**Final Structure** (8 files):
- README.md - Project overview
- STATUS.md - Implementation status
- QUICKSTART.md - 5-minute setup
- SETUP.md - Detailed configuration
- FRIEND_SYSTEM.md - Friend feature docs
- OAUTH_SETUP.md - OAuth configuration
- REALTIME_NOTIFICATIONS.md - SSE architecture
- CHANGELOG.md - Version history

---

## Known Issues

### v1.0.0
- [ ] Test coverage at 15% (needs improvement)
- [ ] Rate limiting not enforced (logic exists)
- [ ] Background cleanup jobs not scheduled
- [ ] No application performance monitoring

---

## Upcoming Features (Roadmap)

### v1.1.0 (Future)
- [ ] Comprehensive test suite
- [ ] CI/CD pipeline
- [ ] Performance monitoring
- [ ] User-created custom questions
- [ ] Game history and statistics
- [ ] Achievements system

### v1.2.0 (Future)
- [ ] Mobile native apps
- [ ] Push notifications
- [ ] User profiles with avatars
- [ ] Multiple game modes
- [ ] Leaderboards

---

## Contributors

This project was built through intensive development sessions focusing on:
- Clean architecture
- Production-ready code
- Comprehensive documentation
- User experience polish

---

## Support & Resources

- **Setup Guide**: See `SETUP.md`
- **Quick Start**: See `QUICKSTART.md`
- **Current Status**: See `STATUS.md`
- **OAuth Setup**: See `OAUTH_SETUP.md`

---

**Last Updated**: November 2025
**Current Version**: 1.0.0
**Status**: ✅ Production Ready
