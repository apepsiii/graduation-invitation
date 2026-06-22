# Project Overview: Undangan Digital SMK NIBA

**Tech Stack:** Go 1.25 + Gin Framework + SQLite + Custom HTML/CSS/JS + OneSender API

---

## Current Progress: ALL PHASES COMPLETE

### Phase 1: Infrastructure (DONE)
- [x] Go module initialized (`undangan-digital`)
- [x] Directory structure: `cmd/`, `internal/` (models, repository, handlers, routes, middleware, installer), `templates/`, `assets/`, `database/`
- [x] SQLite database schema with 5 tables: `guests`, `event_settings`, `guestbooks`, `rundowns`, `galleries`
- [x] Pure Go SQLite driver (modernc.org/sqlite - no CGO required)
- [x] `.env` support with godotenv
- [x] `.env.example` created
- [x] Embedded resources via `go:embed` (templates + assets in binary)

### Phase 2: Backend Core (DONE)
- [x] Gin server entry point (`cmd/main.go`)
- [x] Static file server for `assets/`
- [x] Template engine with custom loader (handles subdirectories)
- [x] All models defined with form tags (`internal/models/models.go`)
- [x] Repository layer with full CRUD operations (`internal/repository/repository.go`)
- [x] Routes registered (`internal/routes/routes.go`)
- [x] CORS middleware
- [x] Session-based auth with HMAC signed cookies (`internal/middleware/session.go`)
- [x] Auto-migration for schema changes (ALTER TABLE)

### Phase 3: Frontend (DONE)
- [x] home.html - Landing page (sage green, modern minimalism)
- [x] invitation.html - Mobile-first sage green design, countdown, QR, RSVP form, guestbook comments
- [x] dashboard.html - Stats and guest table (sage green theme)
- [x] guests.html - CRUD with modal forms, CSV import, broadcast WA
- [x] guestbooks.html - Message management (view/delete)
- [x] settings.html - Event settings + OneSender config (AJAX submission)
- [x] scanner.html - QR scanner with html5-qrcode
- [x] rundowns.html - Rundown management CRUD
- [x] galleries.html - Gallery management CRUD with file upload
- [x] login.html - Custom login form (sage green theme)
- [x] 404.html - Clean modern error page (sage green)
- [x] error.html - Error page template (sage green)

### Phase 4: QR Scanner (DONE)
- [x] QR code generation on invitation page (qrcode.js local, toDataURL)
- [x] Scanner page with camera access
- [x] Scan endpoint validates and marks attendance
- [x] Success/error visual + audio feedback
- [x] Session auth works with fetch requests

### Phase 5: Dashboard Admin (DONE)
- [x] Dashboard with stats (total guests, RSVP, check-ins)
- [x] Guest management (list, add, edit, delete)
- [x] Guest import from CSV (full implementation)
- [x] Event settings CMS
- [x] Rundown management CRUD
- [x] Gallery management CRUD with image upload
- [x] Guestbook/message management (view, delete)
- [x] WhatsApp broadcast via OneSender API (text + image)
- [x] Broadcast test feature
- [x] Broadcast to selected guests or all

### Phase 6: Deployment (DONE)
- [x] Single binary installer with embedded resources
- [x] Interactive wizard (app name, port, instance number)
- [x] YAML config file (`/etc/undangan-digital.yaml`)
- [x] systemd service auto-generation
- [x] Update mechanism (overwrite binary + templates + assets)
- [x] Uninstall feature
- [x] Cross-compile for Linux amd64

### Phase 7: Bracelet (Gelang) Bulk Print (DONE)
- [x] Field `kelas` added to `guests` table (auto-migration via ALTER TABLE)
- [x] CSV import accepts new column `kelas` (format: `name,phone_number,kelas,slug`)
- [x] Form Tambah/Edit Guest updated with `kelas` field
- [x] Bracelet page (`/admin/bracelet`) - 10 gelang per A4 portrait page
- [x] Bracelet dimensions: 2cm × 23cm (vertical, as worn on wrist)
- [x] Server-side QR code generation (base64 PNG embedded in HTML)
- [x] Background image from `assets/uploads/ticket/template.png`
- [x] Content includes: title "8th - 2026 NIBA Graduation", venue/date, school logo (SVG house), student name + kelas, QR code, scan instruction
- [x] Print button → Save as PDF (browser print dialog, scale 100%)
- [x] Guests sorted by kelas then name
- [x] "Perpisahan Ke-10" → "Perpisahan Ke-8" in all templates

### Phase 8: Responsive UI & Broadcast Improvements (DONE)
- [x] Responsive navbar with hamburger menu on ALL admin pages (settings, guests, galleries, guestbooks, rundowns, scanner, meal, dashboard)
- [x] Mobile-friendly navigation: navbar collapses to dropdown on screens < 768px
- [x] Dashboard shortcut: scanner, konsumsi, tamu, cetak gelang buttons
- [x] Clickable stat cards on dashboard (navigate to relevant pages)
- [x] Improved scanner UX: prominent guest name + kelas display, large status icons
- [x] Show attendance timestamp for double check-in detection
- [x] Broadcast template from settings (admin can edit message template)
- [x] Broadcast image URL from settings (used as default for image broadcast)
- [x] Send WhatsApp to single guest from guests table (per-guest button)
- [x] Broadcast status endpoint for real-time progress tracking with polling
- [x] Broadcast logging: per-guest success/failure logged with details
- [x] Maps display: button-only (no iframe preview)
- [x] Date display with day name (Sabtu, 27 Juni 2026) and full time (08:00 - Selesai)
- [x] Bulk download QR codes as ZIP (named by student name)
- [x] Template preview page (`/admin/template-preview`) for design debugging
- [x] Countdown timer on invitation page (live days/hours/minutes/seconds to event)
- [x] "The 10th Graduation Ceremony" → "The 8th Graduation Ceremony" fix

---

## Backend Architecture

```
cmd/main.go                          Entry point (serve mode + installer dispatch)
resources.go                         Embedded templates + assets via go:embed
internal/
  models/models.go                   Struct definitions with form/db tags
  repository/repository.go           Database operations (all CRUD + CSV import + migrations)
  handlers/handlers.go               HTTP handlers (pages + APIs)
  handlers/broadcast.go              OneSender WhatsApp broadcast service
  routes/routes.go                   Route definitions + template loader
  middleware/
    middleware.go                    CORS middleware
    session.go                       Session-based auth (HMAC signed cookies)
  installer/
    installer.go                     VPS installer wizard + systemd setup
    utils.go                         File extraction + copy utilities
templates/
  home.html                          Landing page
  invitation.html                    Guest invitation page
  404.html                           Not found page
  error.html                         Generic error page
  admin/
    login.html                       Admin login
    dashboard.html                   Admin dashboard
    guests.html                      Guest management + broadcast
    guestbooks.html                  Message management
    settings.html                    Event + OneSender settings
    scanner.html                     QR scanner
    rundowns.html                    Rundown CRUD
    galleries.html                   Gallery CRUD with upload
    meal.html                        Meal/Consumption tracking
    bracelet.html                    Bulk print gelang A4 (2cm × 23cm × 10/page)
assets/
  qrcode.min.js                      QR code library (local, not CDN)
  uploads/                           Gallery image uploads
  uploads/ticket/                    Bracelet template assets
    template.png                     Background design for gelang
```

---

## API Endpoints

### Public
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Landing page |
| GET | `/undangan/:slug` | Invitation page for guest |
| POST | `/api/rsvp` | Submit RSVP + message |
| GET | `/api/guestbooks` | Public guestbook comments (paginated, 10/page) |
| GET | `/health` | Health check |

### Auth (Session-based)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/login` | Login page |
| POST | `/admin/login` | Login (form/JSON) |
| GET | `/admin/logout` | Logout (clears session) |

### Admin Pages (Session Protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/dashboard` | Dashboard page |
| GET | `/admin/guests` | Guest management page |
| GET | `/admin/guestbooks` | Message management page |
| GET | `/admin/settings` | Settings page |
| GET | `/admin/scanner` | QR scanner page |
| GET | `/admin/rundowns` | Rundown management page |
| GET | `/admin/galleries` | Gallery management page |
| GET | `/admin/meal` | Meal/Consumption tracking page |
| GET | `/admin/bracelet` | Bulk print gelang A4 (10/page) |

### Admin API (Session Protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/admin/api/settings` | Update event + OneSender settings |
| POST | `/admin/api/guests` | Create guest |
| PUT | `/admin/api/guests` | Update guest |
| DELETE | `/admin/api/guests/:id` | Delete guest |
| POST | `/admin/api/guests/import` | Import guests from CSV |
| POST | `/admin/api/scan` | QR code validation + check-in |
| GET | `/admin/api/stats` | Stats API (JSON) |
| GET | `/admin/api/rundowns` | List rundowns (JSON) |
| POST | `/admin/api/rundowns` | Create rundown |
| DELETE | `/admin/api/rundowns/:id` | Delete rundown |
| GET | `/admin/api/galleries` | List galleries (JSON) |
| POST | `/admin/api/galleries` | Create gallery (file upload or URL) |
| DELETE | `/admin/api/galleries/:id` | Delete gallery |
| DELETE | `/admin/api/guestbooks/:id` | Delete guestbook message |
| POST | `/admin/api/broadcast` | WhatsApp broadcast (text/image) |
| POST | `/admin/api/broadcast/test` | Test broadcast to single number |

---

## Key Features

### 1. Invitation System
- Personal invitation per guest (unique slug)
- QR code for check-in (local library, no CDN)
- Countdown timer
- RSVP form (Hadir / Tidak Hadir / Ragu)
- Guestbook comments with pagination (10 per page)
- 1 message per guest (upsert logic)

### 2. QR Scanner
- Camera-based scanning (html5-qrcode)
- Real-time attendance validation
- Audio + visual feedback
- Session-authenticated

### 3. WhatsApp Broadcast (OneSender)
- Text messages with variables: `{nama}`, `{link}`
- Image messages with caption
- Send to selected guests or all
- Test send before broadcast
- Concurrent sending (3 workers)
- Config: OneSender URL, API Key, App Base URL

### 4. VPS Installer
- Single binary with embedded resources
- Interactive wizard: app name, port, instance
- YAML config at `/etc/undangan-digital.yaml`
- Auto systemd service creation
- Update menu (new binary + templates)
- Uninstall option

### 5. Bracelet (Gelang) Bulk Print
- Halaman: `/admin/bracelet`
- Layout: A4 portrait, 10 gelang per halaman (2cm × 23cm per gelang)
- Orientasi: vertikal (konten di-rotate 90° CCW dari desain horizontal)
- QR code: di-generate server-side sebagai base64 PNG
- Background: `assets/uploads/ticket/template.png`
- Konten gelang: judul event, venue/tanggal, logo sekolah (SVG), nama siswa, kelas, QR, instruksi scan
- Sorting: by kelas → nama
- Output: print ke A4 / Save as PDF (atur printer ke Scale 100%)

### 6. Tim Konsumsi (Meal Tracking)
- Halaman: `/admin/meal`
- Scan QR kedua kali untuk verifikasi makanan
- Reset meal untuk tamu yang salah scan
- Stats: total tamu, sudah ambil makan, sisa
- API: `/admin/api/meal/scan`, `/admin/api/meal/stats`, `/admin/api/meal/checkins`

---

## Database Schema

### `event_settings` (Singleton - 1 row)
- id, event_title, event_date, event_time, venue_name, venue_address, maps_link, dresscode
- onesender_url, onesender_api_key, app_base_url

### `guests`
- id, slug (unique), name, phone_number, kelas, qr_token (unique), rsvp_status, is_attended, attended_at, meal_taken_at, created_at

### `guestbooks`
- id, guest_id (FK), message, created_at
- Constraint: 1 message per guest (enforced in code, dedup on startup)

### `rundowns`
- id, start_time, end_time, activity_name, description

### `galleries`
- id, image_url, caption, sort_order

---

## RSVP Status Constants

```go
RSVPBelumKonfirmasi = "Belum Konfirmasi"
RSVPHadir           = "Hadir"
RSVPTidakHadir      = "Tidak Hadir"
RSVPRagu            = "Ragu"
```

---

## Theme & Design

- **Color:** Sage Green (`#5c6352` primary, `#7d8471` accent, `#e8ebe3` light)
- **Typography:** Cormorant Garamond (serif) + Inter (sans-serif)
- **Style:** Clean modern minimalism elegant
- **Responsive:** Mobile-first design

---

## How to Run

### Development
```bash
cd E:/smk/undangan_digital
go run cmd/main.go serve
```
Default: `http://localhost:8080`
Admin: `http://localhost:8080/admin/dashboard` (Login: admin/admin123)

### Build for VPS (Linux amd64)
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go
```

### VPS Installation
```bash
chmod +x undangan-digital
./undangan-digital
# Follow the wizard...
```

---

## Environment Variables

See `.env.example`:
- DATABASE_PATH
- PORT
- ADMIN_USER
- ADMIN_PASS
- SESSION_SECRET
- BROADCAST_WEBHOOK_URL (legacy, use OneSender settings instead)
- APP_BASE_URL

---

## Handler Struct

```go
type Handler struct {
    repo     *repository.Repository
    session  *middleware.SessionManager
    authUser string
    authPass string
    broadcast *BroadcastService
}
```

`NewHandler(repo, session, authUser, authPass, broadcastService)`

---

## Fixed Issues

1. **Auth Flow** - Session-based auth with HMAC signed cookies
2. **Scanner Auth** - Session cookie automatically sent with fetch (same-origin)
3. **Settings Binding** - Added form tags to models, fixed field names
4. **RSVP Status** - Standardized to "Belum Konfirmasi", "Hadir", "Tidak Hadir", "Ragu"
5. **Template Loading** - Custom loader handles subdirectories correctly
6. **Logout** - Added GET handler for logout link
7. **404 Page** - Created clean modern 404/error pages
8. **Rundown CRUD** - Complete implementation with page and API
9. **Gallery CRUD** - Complete implementation with file upload
10. **CSV Import** - Full implementation with transaction support
11. **Created_at Scanning** - Fixed TEXT to time.Time parsing via sql.NullString
12. **QR Code** - Switched from CDN to local library, toDataURL with retry
13. **Navbar Consistency** - All admin pages have identical navbar (7 links)
14. **Guestbook Dedup** - 1 message per guest, upsert on new submission
15. **Schema Migration** - Auto ALTER TABLE for new columns (onesender fields, meal_taken_at, kelas)
16. **Perpisahan Edition** - "Ke-10" → "Ke-8" updated in all templates
17. **Bracelet Layout** - Server-side QR generation, CSS rotation for vertical wristband, A4 print-friendly pagination
18. **Meal Tracking** - Separate second-scan endpoint for meal consumption (different from check-in)
