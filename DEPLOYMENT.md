# Panduan Deployment VPS

Panduan lengkap untuk deploy Undangan Digital ke VPS (Virtual Private Server).

---

## Prasyarat

- VPS dengan OS Linux (Ubuntu 20.04+ / Debian 10+ / CentOS 8+)
- Akses root atau sudo
- Port yang digunakan terbuka di firewall
- OneSender sudah terinstall (untuk broadcast WhatsApp)

---

## Langkah 1: Build Binary di Local

### Windows (PowerShell/CMD)

```powershell
cd E:\SMK\undangan_digital

# Build untuk Linux amd64
$env:CGO_ENABLED=0
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go

# Reset environment
$env:GOOS="windows"
```

### Linux/Mac

```bash
cd /path/to/undangan_digital
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go
```

---

## Langkah 2: Upload ke VPS

Menggunakan SCP:

```bash
scp undangan-digital root@your-vps-ip:/root/
```

Atau menggunakan SFTP (FileZilla, WinSCP):
- Upload file `undangan-digital` ke `/root/`

---

## Langkah 3: Jalankan Installer

SSH ke VPS:

```bash
ssh root@your-vps-ip
```

Jalankan installer:

```bash
chmod +x undangan-digital
./undangan-digital
```

---

## Langkah 4: Ikuti Wizard

```
╔══════════════════════════════════════════╗
║    Undangan Digital - Installer v1.0    ║
║    Platform Undangan Digital untuk      ║
║    Sekolah & Acara                      ║
╚══════════════════════════════════════════╝

=== INSTALL BARU ===

Nama aplikasi [undangan-digital]: 
Port [8080]: 
Instance keberapa [1]: 
Username admin [admin]: 
Password admin [admin123]: 

=== Konfirmasi ===
Nama aplikasi: undangan-digital
Port: 8080
Instance: 1
Install dir: /opt/undangan-digital
Admin: admin / admin123

Lanjutkan install? [y/N]: y
```

Installer akan:
1. Membuat direktori `/opt/undangan-digital/`
2. Extract templates dan assets
3. Copy binary ke `/usr/local/bin/`
4. Membuat systemd service
5. Enable dan start service
6. Menyimpan config ke `/etc/undangan-digital.yaml`

---

## Langkah 5: Verifikasi

Cek status service:

```bash
systemctl status undangan-digital
```

Cek log:

```bash
journalctl -u undangan-digital -f
```

Test akses:

```bash
curl http://localhost:8080
```

Buka browser: `http://your-vps-ip:8080`

---

## Konfigurasi OneSender (WhatsApp Broadcast)

### 1. Setup OneSender di VPS

Pastikan OneSender sudah berjalan di VPS:

```bash
# Contoh: OneSender di port 3000
curl http://localhost:3000/api/v1/messages
```

### 2. Konfigurasi di Dashboard Admin

Buka: `http://your-vps-ip:8080/admin/settings`

Isi field:
- **OneSender API URL:** `http://localhost:3000` atau `http://127.0.0.1:3000`
- **OneSender API Key:** (dari OneSender dashboard)
- **URL Aplikasi:** `http://your-vps-ip:8080` atau domain `https://undangan.sekolah.sch.id`

Klik **Simpan Konfigurasi**.

### 3. Test Broadcast

Buka: `http://your-vps-ip:8080/admin/guests`

1. Klik **📣 Broadcast WA**
2. Isi pesan: `Halo {nama}, kamu diundang ke acara perpisahan! Link: {link}`
3. Isi nomor test di **Test Kirim**: `628xxxxxxxxxx`
4. Klik **Test Kirim**
5. Cek WhatsApp

---

## Update Aplikasi

### 1. Build versi baru di local

```powershell
# Windows
$env:CGO_ENABLED=0
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go
```

### 2. Upload ke VPS

```bash
scp undangan-digital root@your-vps-ip:/root/
```

### 3. Jalankan update

```bash
cd /root
chmod +x undangan-digital
./undangan-digital
```

Pilih menu **2. Update Aplikasi**

Proses:
- Binary di-replace
- Templates/assets di-refresh
- Service di-restart

---

## Multiple Instance (Opsional)

Jika ingin menjalankan lebih dari 1 instance di VPS yang sama:

```bash
./undangan-digital

# Instance 2
Nama aplikasi: undangan-digital
Port: 8081
Instance keberapa: 2
Install dir: /opt/undangan-digital-2

# Instance 3
Nama aplikasi: undangan-digital
Port: 8082
Instance keberapa: 3
Install dir: /opt/undangan-digital-3
```

Setiap instance punya systemd service terpisah.

---

## Command Service

```bash
# Start
systemctl start undangan-digital

# Stop
systemctl stop undangan-digital

# Restart
systemctl restart undangan-digital

# Status
systemctl status undangan-digital

# Log real-time
journalctl -u undangan-digital -f

# Log 100 baris terakhir
journalctl -u undangan-digital -n 100
```

---

## File Penting

| File | Lokasi | Deskripsi |
|------|--------|-----------|
| Binary | `/usr/local/bin/undangan-digital` | Executable |
| Config | `/etc/undangan-digital.yaml` | Konfigurasi YAML |
| Data | `/opt/undangan-digital/` | Templates, assets, database |
| Database | `/opt/undangan-digital/database/database.sqlite` | SQLite database |
| Uploads | `/opt/undangan-digital/assets/uploads/` | Foto galeri |
| Systemd | `/etc/systemd/system/undangan-digital.service` | Service file |

---

## Backup & Restore

### Backup

```bash
# Backup database dan uploads
tar -czvf undangan-backup-$(date +%Y%m%d).tar.gz \
  /opt/undangan-digital/database \
  /opt/undangan-digital/assets/uploads \
  /etc/undangan-digital.yaml
```

### Restore

```bash
# Extract backup
tar -xzvf undangan-backup-20260617.tar.gz -C /

# Restart service
systemctl restart undangan-digital
```

---

## Uninstall

```bash
./undangan-digital
# Pilih menu 6. Uninstall
# Konfirmasi dengan 'y'
```

Atau manual:

```bash
systemctl stop undangan-digital
systemctl disable undangan-digital
rm /etc/systemd/system/undangan-digital.service
systemctl daemon-reload
rm /usr/local/bin/undangan-digital
rm -rf /opt/undangan-digital
rm /etc/undangan-digital.yaml
```

---

## Troubleshooting

### Port sudah digunakan

```bash
# Cek process di port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Service tidak jalan

```bash
# Cek log
journalctl -u undangan-digital -n 50

# Cek config
cat /etc/undangan-digital.yaml

# Cek file
ls -la /opt/undangan-digital/
```

### Database error

```bash
# Cek database
sqlite3 /opt/undangan-digital/database/database.sqlite ".tables"

# Reset database (HATI-HATI!)
rm /opt/undangan-digital/database/database.sqlite
systemctl restart undangan-digital
```

### Broadcast tidak terkirim

1. Cek OneSender running: `curl http://localhost:3000`
2. Cek API key di `/admin/settings`
3. Cek log: `journalctl -u undangan-digital -f`
4. Test dengan nomor sendiri

---

## SSL/HTTPS (Rekomendasi)

Menggunakan Nginx + Let's Encrypt:

```bash
# Install Nginx
apt install nginx certbot python3-certbot-nginx

# Buat config
nano /etc/nginx/sites-available/undangan

# Isi:
server {
    listen 80;
    server_name undangan.sekolah.sch.id;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Enable
ln -s /etc/nginx/sites-available/undangan /etc/nginx/sites-enabled/
nginx -t
systemctl restart nginx

# SSL
certbot --nginx -d undangan.sekolah.sch.id
```

Update URL aplikasi di `/admin/settings`:
- **URL Aplikasi:** `https://undangan.sekolah.sch.id`

---

## Support

Jika mengalami masalah:
1. Cek log: `journalctl -u undangan-digital -f`
2. Cek status: `systemctl status undangan-digital`
3. Cek config: `cat /etc/undangan-digital.yaml`

---

*Happy deploying! 🚀*