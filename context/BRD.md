Berikut adalah rancangan *Business Requirement Document* (BRD) lengkap untuk proyek aplikasi Anda. Dokumen ini menstrukturkan semua ide dan kebutuhan teknis yang telah kita diskusikan menjadi format standar industri yang siap dijadikan acuan pengembangan.

---

# BUSINESS REQUIREMENT DOCUMENT (BRD)

**Nama Proyek:** Aplikasi Undangan Digital & Sistem Kehadiran Perpisahan SMK NIBA
**Versi Dokumen:** 1.0
**Tanggal:** 10 Juni 2026

## 1. Ringkasan Eksekutif (Executive Summary)

Proyek ini bertujuan untuk membangun sebuah sistem informasi terintegrasi berbasis web untuk mengelola undangan, konfirmasi kehadiran (RSVP), dan pencatatan presensi acara perpisahan sekolah. Sistem ini dirancang untuk menggantikan undangan cetak fisik, memberikan pengalaman digital yang interaktif bagi siswa, serta memudahkan panitia dalam memantau kehadiran dan mengelola informasi acara secara *real-time*.

## 2. Latar Belakang & Tujuan

Pengelolaan acara perpisahan sekolah berskala besar membutuhkan distribusi informasi yang cepat, personal, dan efisien.
**Tujuan Proyek:**

* Menyediakan undangan digital yang dipersonalisasi untuk setiap siswa.
* Mengotomatisasi distribusi undangan melalui WhatsApp.
* Mendigitalisasi proses *check-in* di hari H menggunakan pemindai QR Code untuk mempercepat antrean dan akurasi data.
* Menyediakan *dashboard* terpusat bagi panitia/admin untuk memantau metrik kehadiran secara *live*.

## 3. Ruang Lingkup Proyek (Scope of Work)

Sistem ini terdiri dari dua antarmuka utama:

1. **Frontend (Siswa/Tamu):** Halaman publik berupa undangan digital yang bersifat informatif dan interaktif.
2. **Backend & Dashboard Admin:** Panel kontrol yang dilindungi autentikasi untuk manajemen konten, pemantauan tamu, operasional *scanner*, dan eksekusi pesan massal.

---

## 4. Spesifikasi Fungsional (Functional Requirements)

### 4.1. Modul Frontend (Undangan Digital)

Halaman ini diakses melalui *browser* perangkat seluler oleh tamu undangan.

* **Dynamic Greeting:** Menampilkan nama penerima undangan secara spesifik berdasarkan parameter *slug* pada URL.
* **Informasi Acara:** Menampilkan detail Hari/Tanggal, Waktu, Lokasi (Venue), dan ketentuan pakaian (*Dresscode*).
* **Countdown Timer:** Penghitung waktu mundur interaktif menuju hari pelaksanaan acara.
* **Galeri Foto:** Area penayangan (*view-only*) untuk menampilkan dokumentasi foto kelas/sekolah.
* **Rundown Event:** Daftar urutan acara beserta estimasi waktunya.
* **Integrasi Google Maps:** Peta interaktif yang tertanam (*embedded*) untuk penunjuk arah lokasi.
* **Formulir RSVP & Buku Tamu:** Form bagi siswa untuk mengonfirmasi kehadiran (Hadir/Tidak/Ragu) sekaligus meninggalkan pesan, doa, dan harapan.
* **QR Code Generator:** Sebuah QR Code unik yang ter-generate otomatis di halaman undangan, digunakan sebagai tiket masuk saat acara.

### 4.2. Modul Dashboard Admin

Panel manajemen yang diakses oleh panitia sekolah.

* **Autentikasi Login:** Akses masuk yang diamankan menggunakan *middleware auth* (berbasis *session* atau JWT dari Gin).
* **Manajemen Konten (CMS):** *Form update* untuk mengubah teks informasi acara (Tanggal, Waktu, Rundown, Dresscode, Link Maps) yang akan langsung terefleksi di frontend tanpa perlu menyentuh kode.
* **Monitoring Kehadiran:** *Dashboard* statistik *real-time* yang menampilkan total undangan, jumlah RSVP "Hadir", dan jumlah siswa yang telah *check-in* di lokasi.
* **Buku Tamu Digital:** Tabel yang menampilkan daftar doa, harapan, dan jawaban RSVP yang dikirimkan melalui frontend.
* **Database Siswa:** Fitur untuk mengelola (CRUD) daftar nama siswa dan nomor WhatsApp.

### 4.3. Modul Scanner Kehadiran (Penerima Tamu)

* **Web-Based QR Scanner:** Halaman khusus bagi panitia penerima tamu yang mengakses kamera perangkat (*smartphone/webcam*) untuk memindai QR Code tamu.
* **Validasi Real-time:** Sistem memverifikasi validitas QR Code dan meng-update status "Belum Hadir" menjadi "Hadir" di database beserta cap waktu (*timestamp*).
* **Indikator Visual & Audio:** Memberikan notifikasi sukses (warna hijau/suara *beep*) atau gagal (jika QR tidak valid/sudah di-scan).

### 4.4. Modul Broadcast WhatsApp

* **Custom Broadcast:** Kemampuan mengirim pesan ke ribuan kontak sekaligus dengan menyisipkan variabel dinamis (sapaan nama siswa & tautan URL unik).
* **Background Process:** Eksekusi pengiriman massal tidak mengganggu performa *dashboard* utama.

---

## 5. Spesifikasi Non-Fungsional (Non-Functional Requirements)

* **UI/UX Design:** Desain mengutamakan pendekatan *mobile-first* dan minimalis. Menggunakan palet warna bernuansa pastel (seperti biru lembut, hijau, dan kuning-oranye) untuk memberikan kesan modern, bersih, dan elegan.
* **Performa Tampilan:** Tampilan web dibangun murni menggunakan *Custom HTML* dan CSS (tanpa memuat *framework* komponen berat) untuk menjamin waktu muat halaman (*page load*) di bawah 2 detik pada jaringan 3G/4G standar.
* **Performa Sistem:** Backend harus mampu menangani proses *concurrency*, terutama saat mengeksekusi pengiriman *broadcast* pesan tanpa mengalami *timeout*.

---

## 6. Arsitektur & Teknologi Dasar

| Komponen | Teknologi yang Digunakan | Keterangan |
| --- | --- | --- |
| **Bahasa Pemrograman** | Go (Golang) | Berperan sebagai sistem inti, memproses *logic*, *routing*, dan *concurrency* untuk modul pengiriman pesan massal. |
| **Web Framework** | Gin Web Framework | Menangani *routing* HTTP, merender *file* HTML, dan menyediakan *middleware* keamanan (autentikasi login admin). |
| **Database** | SQLite | Basis data tunggal yang sangat ringan, efisien, dan memadai untuk skala jumlah siswa satu sekolah. |
| **Frontend Utama** | Custom HTML, CSS, Vanilla JS | Struktur antarmuka undangan digital tanpa *bloatware*. |
| **Modul QR Scanner** | HTML5-QRCode (Javascript) | Memungkinkan akses *webcam* atau kamera *smartphone* secara langsung lewat *browser* Chrome/Safari. |
| **Gateway Pengiriman** | WhatsApp API / n8n Webhook | Jalur integrasi pesan massal. Backend Go dapat menembak API secara langsung atau mendelegasikan antrean melalui *workflow* n8n. |