# JavaScript Bundling with esbuild

This document describes the JavaScript bundling system implemented using esbuild via the Go API.

---

## 📋 Overview

The application uses **esbuild** (fastest JavaScript bundler) to concatenate and minify JavaScript files:

- **Development mode:** Individual JS files loaded for easier debugging
- **Production mode:** Minified bundles with external source maps (82-94% size reduction)

### Why esbuild?

- ⚡ **Fastest bundler** - Written in Go, 10-100x faster than webpack
- 🔧 **Go API integration** - No separate Node.js build process needed
- 📦 **Simple configuration** - Minimal setup, works out of the box
- 🗺️ **Source maps** - Both inline (dev) and external (prod) support

---

## 🏗️ Architecture

### Bundle Structure

```
esbuild/app-entry.js → esbuild → static/dist/app.bundle.js (61KB minified)
esbuild/admin-entry.js → esbuild → static/dist/admin.bundle.js (404B minified)
```

### File Organization

```
couple-game/
├── esbuild/                      # Entry point files (build-time only)
│   ├── app-entry.js             # Main app bundle entry
│   └── admin-entry.js           # Admin bundle entry
├── internal/build/               # Build service
│   └── esbuild_service.go       # Bundling service implementation
├── cmd/esbuild/                  # CLI tool
│   └── main.go                  # Build/watch/clean commands
└── static/
    ├── js/                       # Source JavaScript files (runtime)
    │   ├── htmx.min.js
    │   ├── sse.js
    │   ├── ui-utils.js
    │   ├── modal.js
    │   ├── notifications-realtime.js
    │   └── admin.js
    └── dist/                     # Generated bundles (gitignored)
        ├── app.bundle.js
        ├── app.bundle.js.map     # Source map (prod only)
        ├── admin.bundle.js
        └── admin.bundle.js.map   # Source map (prod only)
```

### Security Note

**Entry point files are NOT publicly served:**
- Located in `esbuild/` directory (not in `static/`)
- Build-time artifacts only, read by esbuild to create bundles
- Only generated bundles in `static/dist/` are served at runtime
- Prevents exposing internal build configuration

---

## 📦 Bundles

### App Bundle (app.bundle.js)

Main application bundle loaded on all pages.

**Includes:**
- `htmx.min.js` - HTMX core library
- `sse.js` - HTMX SSE extension
- `ui-utils.js` - Toast, Loading, MobileMenu utilities
- `modal.js` - Modal dialog system
- `notifications-realtime.js` - Real-time notifications

**Size:**
- Development: 358KB (unminified, inline source maps)
- Production: 61KB (minified, 82% reduction)

### Admin Bundle (admin.bundle.js)

Admin-specific bundle loaded only on admin pages.

**Includes:**
- `admin.js` - Admin panel utilities

**Size:**
- Development: 6.7KB (unminified, inline source maps)
- Production: 404B (minified, 94% reduction)

---

## 🛠️ Build Commands

### Makefile Targets

```bash
# Production build (minified + external source maps)
make js-build

# Development build (unminified + inline source maps)
make js-build-dev

# Watch mode (auto-rebuild on changes)
make js-watch

# Clean generated bundles
make js-clean
```

### Direct CLI Usage

```bash
# Production build
ENV=production go run ./cmd/esbuild/main.go build

# Development build
ENV=development go run ./cmd/esbuild/main.go build

# Watch mode
ENV=development go run ./cmd/esbuild/main.go watch

# Clean bundles
go run ./cmd/esbuild/main.go clean
```

---

## 🔧 Configuration

### Build Modes

**Production mode (ENV=production):**
- Minification enabled (MinifySyntax, MinifyWhitespace, MinifyIdentifiers)
- External source maps (separate .js.map files)
- Optimized for size and performance
- Generates `//# sourceMappingURL` comments

**Development mode (ENV=development or unset):**
- No minification (readable output)
- Inline source maps (embedded in bundle)
- Faster builds
- Easier debugging

### esbuild Options

```go
api.Build(api.BuildOptions{
    EntryPoints:       []string{entryPoint},
    Bundle:            true,
    Outfile:           outfile,
    Sourcemap:         mode == Prod ? api.SourceMapExternal : api.SourceMapInline,
    MinifyWhitespace:  mode == Prod,
    MinifyIdentifiers: mode == Prod,
    MinifySyntax:      mode == Prod,
    Target:            api.ES2020,
    Format:            api.FormatIIFE,  // Preserves window.* exports
    LogLevel:          api.LogLevelInfo,
})
```

---

## 🗺️ Source Maps

### What Are Source Maps?

Source maps allow browsers to map minified/bundled code back to the original source files for debugging.

### Production (External Source Maps)

- Separate `.map` files (e.g., `app.bundle.js.map`)
- Only downloaded when DevTools are open
- Smaller bundle size (map not included)
- Better production performance
- Browser auto-detects via `//# sourceMappingURL=app.bundle.js.map`

### Development (Inline Source Maps)

- Embedded in the bundle file
- No separate .map file to track
- Faster builds (no file I/O)
- Larger bundle size (map included)

### Browser Usage

1. Load page → bundle loads (e.g., `app.bundle.js`)
2. Open DevTools → Browser sees `//# sourceMappingURL` comment
3. Browser fetches `.map` file (production only)
4. DevTools displays original unminified code with correct line numbers

### Serving Source Maps

No special server configuration needed:
- Maps served from `/static/dist/*.map`
- Go's `http.FileServer` handles them automatically
- Browser requests maps only when DevTools open

---

## 🔄 Template Loading (Conditional)

Templates conditionally load bundled or individual files based on `ENV` variable.

### Production Templates

```html
{{if eq .Env "production"}}
<!-- Production: Load bundled JavaScript -->
<script src="/static/dist/app.bundle.js"></script>
{{else}}
<!-- Development: Load individual files for debugging -->
<script src="/static/js/htmx.min.js"></script>
<script src="/static/js/sse.js"></script>
<script src="/static/js/ui-utils.js"></script>
<script src="/static/js/notifications-realtime.js" defer></script>
<script src="/static/js/modal.js" defer></script>
{{end}}
```

### Implementation

- `Env` field added to `TemplateData` struct (`internal/handlers/base.go`)
- `RenderTemplate()` populates `Env` from `os.Getenv("ENV")`
- Defaults to `"development"` if not set
- Templates in `templates/partials/layout/head-common.html` and `head-admin.html`

---

## 🚀 Development Workflow

### Full Hot-Reload (3 Terminals)

For the best development experience:

```bash
# Terminal 1: Go server with Air hot-reload
make dev

# Terminal 2: SASS watcher (auto-compile CSS)
make sass-watch

# Terminal 3: JavaScript watcher (auto-bundle JS)
make js-watch
```

**Note:** When you run `make dev`, a colored warning message reminds you to start Terminals 2 and 3.

### One-Time Build

```bash
# Build everything at once
make build

# Run in production mode
ENV=production make run
```

---

## 📐 Implementation Details

### Build Service (internal/build/esbuild_service.go)

**Key methods:**
- `BuildApp()` - Build main app bundle
- `BuildAdmin()` - Build admin bundle
- `BuildAll()` - Build both bundles
- `Watch()` - Watch mode with polling (2-second ticker)
- `Clean()` - Remove generated bundles
- `addSourceMapComment()` - Append `//# sourceMappingURL` to bundles

**Watch mode implementation:**
- Polling-based (esbuild v0.27.1 Go API lacks native watch)
- 2-second ticker checks file modification times
- Rebuilds when changes detected
- Context cancellation for graceful shutdown

### CLI Tool (cmd/esbuild/main.go)

**Commands:**
- `build` - Build bundles once
- `watch` - Watch and rebuild on changes
- `clean` - Remove generated bundles

**Environment handling:**
- Reads `ENV` variable to determine mode
- Defaults to `development` if not set
- Signal handling for graceful shutdown (SIGINT, SIGTERM)

---

## 🐛 Bug Fixes

### showToast() Dependency Fix

**Problem:** `showToast()` was defined in `admin.js` but called by `modal.js` and `notifications-realtime.js`, causing runtime errors on non-admin pages.

**Solution:** Moved `showToast()` to `ui-utils.js` (loaded on all pages):

```javascript
// ui-utils.js (lines 89-102)
function showToast(message, type = 'info') {
    const typeMap = {
        'info': () => Toast.info(message),
        'success': () => Toast.success(message),
        'error': () => Toast.error(message),
        'warning': () => Toast.warning(message)
    };
    if (typeMap[type]) {
        typeMap[type]();
    }
}

// Export globally (line 363)
window.showToast = showToast;
```

---

## 📊 Performance Impact

### Production Improvements

- **HTTP requests:** 5 individual files → 1 bundle (80% reduction)
- **Transfer size:** ~358KB → 61KB (82% reduction for app bundle)
- **Parse time:** Faster due to minification
- **Cache efficiency:** Single bundle = better caching

### Development Experience

- **No change:** Still loads individual files
- **Build time:** +1-2 seconds initial build
- **Debugging:** Same as before (individual files in dev mode)
- **Hot-reload:** Automatic rebuild with `make js-watch`

---

## 🔍 Troubleshooting

### Bundles Not Loading

**Check ENV variable:**
```bash
# In production
ENV=production make build
ENV=production make run

# In development
ENV=development make dev
```

### Source Maps Not Working

**Verify source map URLs:**
```bash
tail -n 3 static/dist/app.bundle.js
# Should show: //# sourceMappingURL=app.bundle.js.map
```

**Verify map files exist:**
```bash
ls -lh static/dist/
# Should see: app.bundle.js.map and admin.bundle.js.map
```

### Watch Mode Not Detecting Changes

**Check file permissions:**
```bash
# Make sure source files are readable
ls -la static/js/
```

**Manually rebuild:**
```bash
make js-build-dev
```

### Build Errors

**Clean and rebuild:**
```bash
make js-clean
make js-build-dev
```

**Check Go dependencies:**
```bash
go mod tidy
```

---

## 📚 References

- **esbuild documentation:** https://esbuild.github.io/
- **esbuild Go API:** https://pkg.go.dev/github.com/evanw/esbuild/pkg/api
- **Source maps spec:** https://sourcemaps.info/spec.html

---

## 🎯 Summary

✅ **Production:** Minified bundles (61KB + 404B) with external source maps
✅ **Development:** Individual files or unminified bundles with inline maps
✅ **Hot-reload:** 3-terminal workflow for full development experience
✅ **Security:** Entry point files not publicly served
✅ **Performance:** 82-94% size reduction in production
✅ **DX:** Colored terminal warnings, automatic rebuilds, easy debugging
