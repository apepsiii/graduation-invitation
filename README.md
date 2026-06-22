# Undangan Digital SMK NIBA

Platform undangan digital untuk acara perpisahan SMK NIBA Business School Bogor. Kirim undangan personal, kelola daftar tamu, dan pantau kehadiran dengan mudah.

## Fitur Utama

- **Undangan Personal** - Setiap tamu mendapat undangan dengan tautan unik dan QR code
- **RSVP Online** - Konfirmasi kehadiran secara digital
- **QR Check-in** - Scan QR code untuk validasi kehadiran
- **Cetak Gelang (Bracelet)** - Generate gelang bulk A4 dengan QR + nama + kelas (10 gelang/halaman)
- **Guestbook** - Ucapan dan doa dari tamu
- **Broadcast WhatsApp** - Kirim undangan via WhatsApp (OneSender API)
- **Dashboard Admin** - Kelola tamu, acara, galeri, dan pesan
- **Single Binary Installer** - Install di VPS dengan satu file

## Tech Stack

- **Backend:** Go 1.25 + Gin Framework
- **Database:** SQLite (modernc.org/sqlite - pure Go, no CGO)
- **Frontend:** Custom HTML/CSS/JS (sage green theme)
- **WhatsApp:** OneSender API integration

## Quick Start

### Development (Local)

```bash
# Clone repository
git clone <repo-url>
cd undangan-digital

# Run server
go run cmd/main.go serve

# Open browser
http://localhost:8080
```

Default login: `admin` / `admin123`

### Build for Production (VPS)

```bash
# Cross-compile for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go

# Upload to VPS
scp undangan-digital root@your-vps:/root/

# On VPS
chmod +x undangan-digital
./undangan-digital
```

## VPS Installation Wizard

📖 **Panduan lengkap deployment:** [DEPLOYMENT.md](./DEPLOYMENT.md)

Saat menjalankan binary di VPS, akan muncul wizard interaktif:

```
╔══════════════════════════════════════════╗
║    Undangan Digital - Installer v1.0    ║
╚══════════════════════════════════════════╝

Menu:
  1. Install Baru
  2. Update Aplikasi
  3. Lihat Status Service
  4. Restart Service
  5. Stop Service
  6. Uninstall
  7. Keluar
```

### Install Baru

Wizard akan meminta:
1. Nama aplikasi (default: `undangan-digital`)
2. Port (default: `8080`)
3. Instance number (untuk multiple deployment)
4. Username admin
5. Password admin

Proses install:
- Extract templates dan assets
- Copy binary ke `/usr/local/bin/`
- Create systemd service
- Enable dan start service
- Config disimpan di `/etc/undangan-digital.yaml`

### Update

Upload binary baru ke VPS, jalankan, pilih menu "Update":
- Binary di-replace
- Templates/assets di-refresh
- Service di-restart

## Konfigurasi OneSender (WhatsApp Broadcast)

1. Buka `/admin/settings`
2. Isi field:
   - **OneSender API URL** - URL server OneSender (contoh: `http://localhost:3000`)
   - **OneSender API Key** - API key dari OneSender
   - **URL Aplikasi** - URL public aplikasi (contoh: `https://undangan.sekolah.sch.id`)

3. Buka `/admin/guests`
4. Klik "📣 Broadcast WA"
5. Isi pesan dengan variabel:
   - `{nama}` - nama tamu
   - `{link}` - link undangan
6. Opsional: isi URL gambar untuk broadcast dengan gambar
7. Test dengan tombol "Test Kirim" atau kirim broadcast

## API Endpoints

### Public
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Landing page |
| `/undangan/:slug` | GET | Invitation page |
| `/api/rsvp` | POST | Submit RSVP |
| `/api/guestbooks` | GET | Guestbook messages |

### Admin
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/login` | GET/POST | Login |
| `/admin/dashboard` | GET | Dashboard |
| `/admin/guests` | GET | Guest management |
| `/admin/bracelet` | GET | Cetak gelang bulk A4 |
| `/admin/settings` | GET | Settings |
| `/admin/scanner` | GET | QR Scanner |
| `/admin/api/broadcast` | POST | WhatsApp broadcast |
| `/admin/api/broadcast/test` | POST | Test broadcast |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_PATH | `database/database.sqlite` | SQLite database path |
| PORT | `8080` | Server port |
| ADMIN_USER | `admin` | Admin username |
| ADMIN_PASS | `admin123` | Admin password |
| SESSION_SECRET | `default-secret` | HMAC signing key |

## Database Schema

- `event_settings` - Konfigurasi acara + OneSender
- `guests` - Data tamu (slug, name, kelas, phone, QR token, RSVP, attendance)
- `guestbooks` - Ucapan/doa dari tamu (1 per tamu)
- `rundowns` - Susunan acara
- `galleries` - Galeri foto

## Theme

- **Primary Color:** Sage Green (#5c6352)
- **Accent:** #7d8471
- **Light:** #e8ebe3
- **Typography:** Cormorant Garamond + Inter
- **Style:** Modern minimalism elegant

## Directory Structure

```
undangan-digital/
├── cmd/main.go              # Entry point
├── resources.go             # Embedded files
├── internal/
│   ├── models/              # Data models
│   ├── repository/          # Database layer
│   ├── handlers/            # HTTP handlers + broadcast
│   ├── routes/              # Route definitions
│   ├── middleware/          # Session auth
│   └── installer/           # VPS installer
├── templates/               # HTML templates
│   ├── home.html
│   ├── invitation.html
│   ├── 404.html
│   └── admin/
│       ├── login.html
│       ├── dashboard.html
│       ├── guests.html
│       ├── guestbooks.html
│       ├── settings.html
│       ├── scanner.html
│   ├── rundowns.html
│   ├── galleries.html
│   └── bracelet.html                # Cetak gelang bulk A4
├── assets/
│   ├── qrcode.min.js
│   └── uploads/
│       └── ticket/
│           └── template.png         # Template gelang (background)
├── database/
├── AGENTS.md
├── README.md
├── go.mod
└── go.sum
```

## License

MIT License - SMK NIBA Business School Bogor

## Author

Developed for SMK NIBA Business School Bogor - Perpisahan Ke-8 (2026)