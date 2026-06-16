Berikut adalah rancangan fase pengembangan (Roadmap) dari awal hingga akhir. Panduan ini disusun dengan pendekatan langkah demi langkah agar proyek tetap terorganisir, *maintainable*, dan siap diluncurkan tepat waktu untuk acara perpisahan di sekolah.

---

### Phase 1: Inisiasi & Persiapan Infrastruktur (Minggu 1)

Fase ini fokus pada fondasi proyek agar struktur folder dan *database* tertata dengan benar sejak hari pertama.

* **1. Setup Repositori & Struktur Direktori:**
* Inisialisasi proyek Go (`go mod init`).
* Buat kerangka direktori standar: `cmd/`, `internal/` (models, repository, handlers, routes, middleware), `templates/`, dan `assets/`.


* **2. Desain & Inisialisasi Database SQLite:**
* Buat *file* `database.sqlite`.
* Tulis *script* SQL untuk mengeksekusi pembuatan tabel (seperti yang ada di PRD: `guests`, `event_settings`, `guestbooks`, `rundowns`, `galleries`).
* Buat koneksi *database* di Go menggunakan *driver* SQLite.


* **3. Konfigurasi Environment:**
* Buat *file* `.env` untuk menyimpan konfigurasi (PORT server, kredensial Admin, *secret key* JWT/Session, dan URL *Webhook* WhatsApp API).



### Phase 2: Pengembangan Backend Core & Routing (Minggu 2)

Membangun "mesin" dari aplikasi menggunakan Gin Framework.

* **1. Setup Gin Server:**
* Konfigurasi *entry point* di `cmd/main.go` untuk menjalankan server Gin.
* Daftarkan *static file server* untuk folder `assets/` (CSS, JS, Gambar).
* Daftarkan *template engine* (`router.LoadHTMLGlob("templates/*")`).


* **2. Pembuatan Model & Repository:**
* Buat *struct* Go di folder `internal/models/` yang merepresentasikan struktur tabel.
* Tulis fungsi CRUD dasar di `internal/repository/` (misalnya: fungsi `GetGuestBySlug`, `UpdateRSVP`, `InsertGuestbook`).


* **3. Setup Routing & Middleware Awal:**
* Buat kerangka *endpoint* publik (contoh: `/undangan/:slug`).
* Implementasikan *middleware* autentikasi (BasicAuth atau JWT) untuk memproteksi *route* `/admin/*`.



### Phase 3: Pengerjaan Frontend "Mobile-First" (Minggu 3)

Fase ini berfokus pada visual undangan yang akan dilihat oleh siswa. Penggunaan *Custom HTML* dan Vanilla CSS murni akan memastikan web termuat secepat kilat.

* **1. Desain UI Undangan (Custom HTML):**
* Rancang tata letak *mobile-first* yang minimalis.
* Terapkan palet warna pastel (seperti nuansa biru lembut, hijau, atau kuning-oranye) melalui CSS kustom untuk memberikan kesan perpisahan yang hangat dan elegan tanpa *bloatware*.


* **2. Integrasi Data ke Template Gin:**
* Hubungkan *handler* Go agar mengirimkan data dari SQLite (Nama Tamu, Rundown, Event Settings) ke dalam variabel *template* HTML.
* Buat logika *Countdown Timer* menggunakan Vanilla Javascript di sisi *client*.


* **3. Pembuatan Formulir RSVP & Buku Tamu:**
* Buat *form* HTML untuk konfirmasi kehadiran dan pengisian doa/pesan.
* Tulis fungsi *Fetch* di Javascript untuk mengirim data secara *asynchronous* ke *endpoint* API POST `/api/rsvp`.



### Phase 4: Integrasi Fitur Krusial & Scanner Kehadiran (Minggu 4)

Menyelesaikan fitur interaktif untuk hari H.

* **1. Pembuatan QR Code Siswa:**
* Buat sistem yang me- *render* QR Code di halaman undangan siswa berdasarkan kolom `qr_token` dari database.


* **2. Halaman Web Scanner Admin:**
* Buat halaman HTML khusus admin/penerima tamu.
* Integrasikan *library* `html5-qrcode.js` untuk mengaktifkan akses kamera *smartphone*.


* **3. Logic Penerimaan Scan:**
* Buat *endpoint* di Go (`POST /api/admin/scan`) yang menerima hasil pemindaian.
* Pastikan sistem melakukan validasi: jika berhasil, update status `is_attended` di SQLite dan kirim respons sukses (berwarna hijau & suara *beep*) ke layar penerima tamu.



### Phase 5: Dashboard Admin & Automasi WhatsApp (Minggu 5)

Fase pengelolaan data dan distribusi informasi secara massal.

* **1. Antarmuka Dashboard Admin:**
* Buat halaman tabel daftar siswa, status RSVP, dan manajemen CMS (update tanggal, lokasi).
* Buat halaman pemantauan statistik kehadiran *real-time*.


* **2. Sistem Broadcast Background (Goroutines):**
* Siapkan *handler* Go untuk tombol "Kirim Undangan WA".
* Tulis logika *concurrency* (*Goroutine*) agar sistem melakukan *looping* pengiriman pesan massal di *background* tanpa membuat server *crash*.


* **3. Integrasi Gateway WhatsApp:**
* Hubungkan *script* Go agar menembak *webhook* ke *platform* automasi (seperti n8n) yang kemudian memproses antrean pesan ke API WhatsApp dengan variabel dinamis (Sapaan Nama & Tautan *Slug*).



### Phase 6: Testing, Deployment & Gladi Resik (Minggu 6)

Tahap final sebelum aplikasi resmi digunakan oleh ratusan pengguna.

* **1. End-to-End Testing:**
* Uji beban (*load test*) pada fitur *broadcast* untuk memastikan *Goroutines* berjalan aman.
* Uji respons UI pada berbagai ukuran layar *smartphone*.
* Uji pemindaian QR menggunakan kamera *smartphone* dengan kondisi cahaya terang dan redup.


* **2. Deployment:**
* *Compile* kode Go menjadi *binary file* eksekusi.
* Unggah *binary file*, folder `assets/`, `templates/`, `.env`, dan `database.sqlite` ke *Virtual Private Server* (VPS).
* Konfigurasi *domain* dan sertifikat SSL (HTTPS) agar akses *kamera* di web *scanner* dapat diizinkan oleh *browser*.


* **3. Simulasi Lapangan (Gladi Resik):**
* Lakukan simulasi di lokasi acara (*venue*) bersama panitia penerima tamu untuk memastikan koneksi internet dan *flow* penggunaan *scanner* lancar.