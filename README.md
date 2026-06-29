# Undangan Digital SMK NIBA

Platform undangan digital untuk acara perpisahan SMK NIBA Business School Bogor. Kirim undangan personal, kelola daftar tamu, dan pantau kehadiran dengan mudah.

## Fitur Utama

- **Undangan Personal** - Setiap tamu mendapat undangan dengan tautan unik dan QR code
- **RSVP Online** - Konfirmasi kehadiran secara digital
- **QR Check-in** - Scan QR code untuk validasi kehadiran dengan timestamp
- **Scan Konsumsi** - Scan kedua untuk verifikasi pengambilan makan
- **Cetak Gelang (Bracelet)** - Generate gelang bulk A4 dengan QR + nama + kelas (10 gelang/halaman)
- **Bulk Download QR** - Download semua QR code sebagai ZIP (nama file = nama siswa)
- **Guestbook** - Ucapan dan doa dari tamu
- **Broadcast WhatsApp** - Kirim undangan via WhatsApp (OneSender API)
- **Kirim WA Individual** - Kirim WhatsApp ke satu tamu直接从 tabel
- **Dashboard Admin** - Kelola tamu, acara, galeri, pesan, dan statistics
- **Responsive Mobile** - Navigasi hamburger menu untuk semua halaman admin
- **Single Binary Installer** - Install di VPS dengan satu file

## Tech Stack

- **Backend:** Go 1.25 + Gin Framework
- **Database:** SQLite (modernc.org/sqlite - pure Go, no CGO)
- **Frontend:** Custom HTML/CSS/JS (sage green theme, mobile-first)
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

## Konfigurasi

### OneSender WhatsApp

1. Buka `/admin/settings`
2. Isi field:
   - **OneSender API URL** - URL server OneSender
   - **OneSender API Key** - API key dari OneSender
   - **URL Aplikasi** - URL public aplikasi

### Template Broadcast

1. Buka `/admin/settings`
2. Di section "Template Pesan WhatsApp":
   - **Template Pesan** - template dengan variabel `{nama}` dan `{link}`
   - **URL Gambar** - opsional, untuk broadcast gambar

### Broadcast ke Tamu

1. Buka `/admin/guests`
2. Klik **"📣 Broadcast WA"** untuk broadcast ke semua atau terpilih
3. Atau klik **"Kirim WA"** per tamu langsung dari tabel
4. Variabel template: `{nama}`, `{link}`

## Halaman Admin

| Halaman | Fungsi |
|---------|--------|
| `/admin/dashboard` | Overview stats, shortcut scanner/konsumsi/tamu/gelang |
| `/admin/guests` | Manajemen tamu, import CSV, broadcast WA, kirim WA individual, download QR |
| `/admin/scanner` | Scan QR check-in dengan riwayat |
| `/admin/meal` | Scan QR konsumsi dengan riwayat |
| `/admin/bracelet` | Cetak gelang bulk A4 (10/halaman) |
| `/admin/guestbooks` | Lihat dan hapus ucapan tamu |
| `/admin/rundowns` | Manajemen rundown acara |
| `/admin/galleries` | Manajemen galeri foto |
| `/admin/settings` | Pengaturan acara, OneSender, template broadcast |

## API Endpoints

### Public
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Landing page |
| `/undangan/:slug` | GET | Invitation page |
| `/api/rsvp` | POST | Submit RSVP |
| `/api/guestbooks` | GET | Guestbook messages |

### Admin API
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/api/guests` | POST | Create guest |
| `/admin/api/guests` | PUT | Update guest |
| `/admin/api/guests/:id` | DELETE | Delete guest |
| `/admin/api/guests/import` | POST | Import CSV |
| `/admin/api/guests/qrcodes` | GET | Download all QR as ZIP |
| `/admin/api/scan` | POST | QR check-in scan |
| `/admin/api/meal/scan` | POST | QR meal scan |
| `/admin/api/attendance/checkins` | GET | Attendance history |
| `/admin/api/attendance/:id` | DELETE | Reset attendance |
| `/admin/api/broadcast` | POST | Broadcast WhatsApp |
| `/admin/api/broadcast/single` | POST | Send to single guest |
| `/admin/api/broadcast/test` | POST | Test broadcast |
| `/admin/api/stats` | GET | Dashboard stats |

## Database Schema

- `event_settings` - Konfigurasi acara + OneSender + template broadcast
- `guests` - Data tamu (slug, name, kelas, phone, QR token, RSVP, attendance, meal)
- `guestbooks` - Ucapan/doa dari tamu (1 per tamu)
- `rundowns` - Susunan acara
- `galleries` - Galeri foto

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_PATH | `database/database.sqlite` | SQLite database path |
| PORT | `8080` | Server port |
| ADMIN_USER | `admin` | Admin username |
| ADMIN_PASS | `admin123` | Admin password |
| SESSION_SECRET | `default-secret` | HMAC signing key |

## Theme

- **Primary Color:** Sage Green (#5c6352)
- **Accent:** #7d8471
- **Light:** #e8ebe3
- **Typography:** Cormorant Garamond + Inter
- **Style:** Modern minimalism elegant
- **Responsive:** Mobile-first design dengan hamburger menu

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
├── templates/
│   ├── home.html           # Landing page
│   ├── invitation.html     # Guest invitation
│   ├── 404.html
│   └── admin/
│       ├── login.html
│       ├── dashboard.html   # Stats + shortcuts
│       ├── guests.html      # CRUD + broadcast + WA individual
│       ├── scanner.html     # Check-in scan + history
│       ├── meal.html        # Konsumsi scan + history
│       ├── bracelet.html    # Bulk print gelang
│       ├── guestbooks.html
│       ├── rundowns.html
│       ├── galleries.html
│       └── settings.html    # Event + OneSender + template
├── assets/
│   ├── qrcode.min.js
│   └── uploads/
│       └── ticket/
│           └── template.png # Template gelang background
├── database/
├── release/                 # Compiled binaries
├── AGENTS.md
├── README.md
├── DEPLOYMENT.md
├── go.mod
└── go.sum
```

## License

MIT License - SMK NIBA Business School Bogor

## Author

Developed for SMK NIBA Business School Bogor - Perpisahan Ke-8 (2026)
