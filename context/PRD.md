Berikut adalah **Product Requirements Document (PRD)** yang komprehensif untuk aplikasi Undangan Digital Perpisahan Sekolah Anda, lengkap dengan struktur basis data (Database Schema) menggunakan SQLite.

---

# PRODUCT REQUIREMENTS DOCUMENT (PRD)

**Nama Produk:** Sistem Undangan & Manajemen Kehadiran Perpisahan Sekolah
**Platform:** Web Application (Mobile-First)
**Tech Stack:** Go (Gin Framework), SQLite, Custom HTML/CSS/JS

## 1. Pendahuluan

**1.1 Visi Produk**
Menciptakan pengalaman perpisahan sekolah yang modern dan efisien melalui sistem undangan digital terintegrasi yang memudahkan distribusi, interaksi siswa, dan manajemen presensi acara secara *real-time*.

**1.2 Target Pengguna (User Personas)**

* **Siswa/Tamu Undangan:** Menginginkan informasi acara yang jelas, akses mudah dari *smartphone*, dan antarmuka yang menarik.
* **Panitia/Admin Acara:** Membutuhkan alat yang cepat untuk menyebarkan undangan massal, mengubah detail acara, dan memonitor kehadiran tanpa proses manual yang rumit.
* **Penerima Tamu (Front Desk):** Membutuhkan alat pemindai (scanner) yang responsif dan anti-lelet di pintu masuk acara.

## 2. User Stories

### Frontend (Tamu)

* *Sebagai tamu*, saya ingin melihat undangan dengan nama saya tercantum agar terasa personal.
* *Sebagai tamu*, saya ingin melihat detail acara (waktu, lokasi, dresscode) dan *countdown* agar tidak salah jadwal.
* *Sebagai tamu*, saya ingin mengirim konfirmasi kehadiran (RSVP) dan pesan perpisahan langsung dari web.
* *Sebagai tamu*, saya ingin mendapatkan QR Code unik sebagai tiket masuk acara.

### Backend (Admin/Panitia)

* *Sebagai admin*, saya ingin memperbarui informasi acara dari *dashboard* tanpa harus mengubah kode.
* *Sebagai admin*, saya ingin mengirim pesan WhatsApp massal yang berisi tautan unik ke masing-masing siswa dengan satu kali klik.
* *Sebagai admin*, saya ingin melihat statistik RSVP dan daftar kehadiran secara *real-time*.
* *Sebagai penerima tamu*, saya ingin memindai QR Code tamu menggunakan kamera HP saya agar proses *check-in* lebih cepat.

---

## 3. Spesifikasi Fitur Utama

### 3.1. Frontend Undangan (Publik, dengan Slug)

* **Dynamic Routing & Templating:** Di-render oleh Gin menggunakan HTML template. URL berformat `/undangan/:slug`.
* **Komponen UI:**
* Greeting dinamis (contoh: "Kepada Yth. Budi Santoso").
* Countdown Timer (menghitung mundur ke waktu acara).
* Detail Acara & Google Maps Embed.
* Rundown List & Photo Gallery (Grid View).
* Form RSVP & Kirim Pesan (POST request ke API).
* QR Code Container (Di-generate di *server* atau via JS menggunakan token dari *database*).



### 3.2. Dashboard Admin (Restricted - Gin BasicAuth/JWT)

* **CMS Detail Acara:** Form untuk melakukan Update data ke tabel `event_settings`.
* **Manajemen Siswa & Broadcast:** Tabel daftar siswa dengan tombol "Kirim Undangan WA". Proses ini memicu *Goroutines* di Go untuk mengirim HTTP request ke *gateway* WhatsApp API.
* **Buku Tamu & Statistik:** Penampil data (Tabel RSVP dan Pesan), serta metrik jumlah total undangan, jumlah konfirmasi Hadir, dan jumlah *check-in* aktual.

### 3.3. Web QR Scanner

* **Interface:** Halaman web menggunakan pustaka `html5-qrcode`.
* **Logic:** Kamera membaca QR (berisi `qr_token`). JS melakukan *Fetch* POST ke `/api/admin/scan`. Server memvalidasi token, mengubah status kehadiran, dan membalas dengan JSON `{"status": "success", "guest_name": "Budi Santoso"}`.
* **Feedback:** UI menampilkan centang hijau dan memainkan efek suara.

---

## 4. Struktur Database (SQLite Schema)

Karena kita menggunakan SQLite dan Go, skema dirancang agar relasional, ringan, dan cepat untuk operasi baca/tulis sederhana. Terdapat 5 tabel utama.

### 4.1. Tabel `event_settings`

Menyimpan konfigurasi web undangan. Hanya berisi 1 baris data (Singleton) yang di-update oleh admin.

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER (PK) | Primary Key (selalu 1) |
| `event_title` | TEXT | Nama Acara |
| `event_date` | TEXT | Tanggal acara (Format: YYYY-MM-DD) |
| `event_time` | TEXT | Waktu acara (contoh: 08:00 - Selesai) |
| `venue_name` | TEXT | Nama tempat |
| `venue_address` | TEXT | Alamat lengkap |
| `maps_link` | TEXT | URL Google Maps / iframe embed |
| `dresscode` | TEXT | Ketentuan pakaian |

### 4.2. Tabel `guests`

Tabel utama untuk daftar siswa/tamu undangan.

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER (PK) | Auto Increment |
| `slug` | TEXT (Unique) | Identifier URL (contoh: `budi-santoso-123`) |
| `name` | TEXT | Nama lengkap tamu |
| `phone_number` | TEXT | Nomor WA untuk Broadcast |
| `qr_token` | TEXT (Unique) | String unik (UUID/Hash) untuk di-encode ke QR |
| `rsvp_status` | TEXT | Pilihan: `Belum Konfirmasi`, `Hadir`, `Tidak Hadir` |
| `is_attended` | BOOLEAN | Status Check-in Hari H (`0` = Belum, `1` = Sudah) |
| `attended_at` | DATETIME | Waktu scan QR Code (NULL jika belum hadir) |

### 4.3. Tabel `guestbooks`

Menyimpan pesan/doa yang dikirimkan melalui form frontend.

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER (PK) | Auto Increment |
| `guest_id` | INTEGER (FK) | Relasi ke tabel `guests` |
| `message` | TEXT | Isi harapan/doa |
| `created_at` | DATETIME | Waktu pesan dikirim |

### 4.4. Tabel `rundowns`

Menyimpan jadwal susunan acara yang bisa diubah oleh admin.

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER (PK) | Auto Increment |
| `start_time` | TEXT | Waktu mulai (contoh: 08:00) |
| `end_time` | TEXT | Waktu selesai (contoh: 09:00) |
| `activity_name` | TEXT | Nama kegiatan (contoh: Pembukaan) |
| `description` | TEXT | Keterangan opsional |

### 4.5. Tabel `galleries`

Menyimpan *link* atau *path* file foto untuk ditampilkan di frontend.

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER (PK) | Auto Increment |
| `image_url` | TEXT | Path lokasi gambar statis (contoh: `/assets/img1.jpg`) |
| `caption` | TEXT | Judul/teks gambar (opsional) |
| `sort_order` | INTEGER | Urutan gambar saat ditampilkan |

---

## 5. Alur Data (Data Flow) & Endpoints (Gin)

* **Maintainability (Kemudahan Pemeliharaan):** Kode harus ditulis dengan prinsip *Clean Code* dan *Separation of Concerns* (pemisahan tugas). Logika bisnis, akses *database*, dan *routing* HTTP tidak boleh dicampur dalam satu *file* atau fungsi yang sama.
* **Konfigurasi Dinamis:** Konfigurasi yang bersifat sensitif atau sering berubah (seperti *port* server, *secret key* JWT, URL *webhook* n8n/WhatsApp API, dan *path database* SQLite) harus disimpan di dalam *file* `.env` dan tidak di-*hardcode* ke dalam *source code*.
* **Dokumentasi Kode:** Setiap fungsi utama, *middleware*, dan *struct* harus memiliki komentar penjelasan (*docstrings*) yang jelas untuk memudahkan *developer* lain memahami alur kerja aplikasi.

**Frontend Routes:**

* `GET /undangan/:slug` $\rightarrow$ Mengambil data dari `guests`, `event_settings`, `rundowns`, `galleries` dan me-render HTML.
* `POST /api/rsvp` $\rightarrow$ Menerima payload (guest_id, rsvp_status, message). Update tabel `guests` dan Insert ke tabel `guestbooks`.

**Admin Routes (Protected by Middleware):**

* `GET /admin/dashboard` $\rightarrow$ Render UI Dashboard admin.
* `POST /api/admin/broadcast` $\rightarrow$ Trigger *Goroutine* loop untuk tabel `guests`.
* `POST /api/admin/scan` $\rightarrow$ Menerima JSON `{"qr_token": "..."}`. Update `guests.is_attended = 1` dan `guests.attended_at = NOW()`.
* CRUD Endpoints untuk `guests`, `rundowns`, dan `event_settings`.


## 6. Kriteria Penerimaan (Acceptance Criteria)

1. **Page Load:** Halaman undangan harus dimuat kurang dari 2 detik (mendukung CSS kustom seefisien mungkin).
2. **Concurrency:** Proses *Broadcast WhatsApp* kepada ratusan data tidak boleh memicu *Timeout/Error 500* pada *dashboard* admin.
3. **QR Scanner:** Modul scanner harus dapat menangkap QR Code pada kondisi cahaya ruangan normal dalam waktu maksimal 2 detik.
4. **Keamanan:** URL *Dashboard* admin dan seluruh API admin menolak akses masuk jika tidak menyertakan *header authorization* atau *session* yang valid (Response 401 Unauthorized).

## Bab 7. Standar Struktur Proyek (Project Structure)

Untuk menjamin skalabilitas dan kerapian, proyek ini akan mengadopsi standar tata letak proyek Go yang umum digunakan komunitas (*Standard Go Project Layout*), disesuaikan dengan kebutuhan *framework* Gin.

**Struktur Direktori Utama:**

* **`cmd/`**
Berisi *entry point* utama aplikasi. File `main.go` di sini hanya bertugas melakukan inisialisasi server, membaca konfigurasi `.env`, dan memanggil *router* utama.
* **`internal/`**
Tempat menyimpan seluruh kode yang bersifat *private* dan merupakan inti logika aplikasi. Direktori ini akan dibagi lagi menjadi:
* **`models/`**: Definisi *struct* Go yang merepresentasikan skema tabel SQLite.
* **`repository/`**: Fungsi-fungsi yang berinteraksi langsung dengan *database* (operasi CRUD menggunakan `database/sql` atau ORM).
* **`handlers/`**: Fungsi pengontrol yang menerima *request* HTTP dari Gin, memanggil logika bisnis, dan mengembalikan *response* JSON atau merender HTML.
* **`routes/`**: Tempat mendaftarkan semua *endpoint* URL, grup API, dan menyematkan *middleware* autentikasi.
* **`middleware/`**: Logika pencegatan *request* seperti sistem *Login/Auth* dan pengecekan *session* admin.


* **`templates/`**
Tempat menyimpan seluruh *file* Custom HTML (*view*) yang akan di-render oleh Gin.
* **`assets/`** (atau `static/`)
Menyimpan aset statis publik seperti *file* CSS, Vanilla Javascript, gambar galeri, dan *library* pihak ketiga seperti `html5-qrcode.js`.
* **`pkg/`** (Opsional)
Menyimpan kode utilitas atau fungsi *helper* yang bisa digunakan ulang di berbagai bagian aplikasi (misalnya fungsi *generator* token acak atau format tanggal).
* **`database/`**
Lokasi penyimpanan *file* fisik `database.sqlite` dan skrip migrasi awal.

---