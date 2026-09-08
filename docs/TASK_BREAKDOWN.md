# Task Breakdown - Article Service

Acuan: [PRD](PRD.md). Checklist ini mencakup implementasi dan verifikasi; belum menandakan pekerjaan sudah dijalankan. Docker adalah kebutuhan proyek yang dipilih, sedangkan k6 tetap opsional. Detail di luar assessment mengikuti asumsi yang ditandai dalam PRD.

## Kondisi Awal

- [x] Proyek Go dengan Gin tersedia.
- [x] Endpoint `GET /health` dan file tesnya tersedia.
- [x] PRD tersedia, termasuk keputusan search, author filtering, cursor pagination, dan Docker.
- [ ] Endpoint artikel, PostgreSQL, migrations, dan konfigurasi Docker belum diimplementasikan.

## T01 - Tetapkan Kontrak API

Dependensi: tidak ada. Acuan: PRD 4 dan 6.

- [x] Dokumentasikan request create: `author_id`, `title`, dan `body`; author harus sudah tersedia.
- [x] Tetapkan format respons artikel, daftar dengan next cursor, dan error JSON; gunakan status HTTP sesuai PRD.
- [x] Dokumentasikan validasi field kosong, JSON tidak valid, author tidak ditemukan, serta limit/cursor tidak valid.
- [x] Tetapkan format ID, timestamp, encoding cursor, default limit, dan maksimum limit sebagai asumsi implementasi.
- [x] Catat aturan: semua kata query harus ditemukan di title/body, nama author cocok penuh tanpa membedakan kapitalisasi, filter gabungan menggunakan AND, dan parameter kosong dianggap tidak diberikan.

Selesai jika: tersedia contoh request/response dan aturan input yang bisa langsung dijadikan acuan implementasi dan tes.

Status: selesai. Hasil: [API Contract](API_CONTRACT.md), dengan ringkasan PRD yang sudah diselaraskan. Verifikasi dokumentasi lulus: delapan contoh JSON valid, format UUID/timestamp sesuai, cursor berhasil decode/encode ulang dan cocok dengan artikel terakhir, serta contoh pagination timestamp sama dan hasil kosong konsisten. Endpoint dan tes executable belum diimplementasikan pada T01.

## T02 - Susun Layer dan Bootstrap

Dependensi: T01. Acuan: PRD 5 dan 7.

- [x] Susun `cmd/`, `api/controller/`, `api/route/`, `bootstrap/`, `domain/`, `repository/`, `usecase/`, dan `migrations/` saat mulai digunakan.
- [x] Gunakan alur Route → Controller → Usecase → Repository → Database dengan dependency wiring langsung.
- [x] Definisikan model Article/Author dan kontrak yang diperlukan; HTTP tetap di controller, SQL di repository.
- [x] Muat konfigurasi environment dan buat satu connection pool PostgreSQL yang dipakai bersama; tangani kegagalan startup dan tutup resource saat shutdown.
- [x] Tambahkan driver SQL PostgreSQL, rapikan dependensi yang tidak dipakai, dan pertahankan health endpoint beserta tesnya.

Selesai jika: aplikasi dapat dijalankan dengan konfigurasi database dan health test tetap lulus setelah perubahan struktur.

Status: selesai. Route dan controller menangani health; bootstrap memuat `DATABASE_URL`/`ADDRESS` lalu membuka satu pool PostgreSQL. `domain` menyediakan model awal; folder layer artikel dibuat saat T05/T06 agar tidak menjadi scaffold kosong.

## T03 - Schema, Migrations, dan Seed Author

Dependensi: T01. Acuan: PRD 5 dan 6.

- [x] Buat SQL migrations untuk `authors(id text, name text)` dan `articles(id text, author_id text, title text, body text, created_at timestamp)`.
- [x] Tambahkan primary key dan foreign key `articles.author_id → authors.id`; jangan menganggap nama author unik.
- [x] Terapkan pembuatan ID/timestamp sesuai kontrak serta constraint untuk field yang wajib diisi.
- [x] Sediakan seed author dengan ID yang terdokumentasi dan bisa dijalankan ulang tanpa menggandakan author seed.
- [x] Dokumentasikan urutan menjalankan migrations dan seed pada database kosong maupun database yang sudah berisi data.

Selesai jika: schema dapat disiapkan ulang secara terkontrol, author seed tersedia, dan database menolak referensi author yang tidak valid.

Status: selesai. Schema dan seed ada di `migrations/`; langkah penerapan/reset dan ID author seed ada di [Database Setup](DATABASE_SETUP.md). PostgreSQL melalui Compose berhasil menerapkan migration 1–3, membuat dua author seed, dan rerun migrator menghasilkan `no change`. Search/index migrations tetap menjadi T06.

## T04 - Docker dan Setup Lokal

Dependensi: T02 dan T03. Acuan: PRD 5 dan 11.

- [x] Buat `Dockerfile` untuk membangun dan menjalankan API serta `.dockerignore` untuk mengecualikan file yang tidak diperlukan dan rahasia lokal.
- [x] Buat `compose.yaml` dengan service `api` dan `db`, koneksi antarservice, serta named volume PostgreSQL.
- [x] Tambahkan database healthcheck dan dependency `service_healthy` sebelum API dimulai.
- [x] Sediakan `.env.example`; pastikan kredensial nyata tidak masuk image atau Git.
- [x] Dokumentasikan konfigurasi, persiapan schema/seed, `docker compose up --build`, dan shutdown. Jangan bergantung pada penghapusan volume untuk menerapkan migrations berikutnya.

Selesai jika: API dan database dapat dijalankan lewat Compose mengikuti petunjuk setup, dan endpoint health dapat diakses.

Status: selesai. Compose menjalankan `db`, migrator satu-kali, dan `api`; database sehat, migration/seed sukses, rerun migrator aman, dan `GET /health` mengembalikan `200`.

## T05 - Implementasi Create Article

Dependensi: T01-T04. Acuan: PRD 4 dan 6.

- [ ] Hubungkan `POST /articles` dari route sampai repository.
- [ ] Validasi input dan petakan author tidak ditemukan menjadi error sesuai kontrak.
- [ ] Simpan artikel dengan SQL berparameter dan satu INSERT atomic; kirim `201` setelah penyimpanan berhasil.
- [ ] Teruskan request context hingga operasi database dan gunakan data request lokal.
- [ ] Tambahkan tes create sukses, persistensi data, input tidak valid, dan author tidak ditemukan.

Selesai jika: artikel yang berhasil dibuat tersimpan dengan ID/timestamp dan respons sesuai kontrak; request tidak valid tidak membuat artikel.

## T06 - List, Search, Filter, dan Pagination

Dependensi: T05. Acuan: PRD 4 dan 8.

- [ ] Implementasikan `GET /articles` dengan urutan `created_at DESC, id DESC` dan daftar kosong saat tidak ada hasil.
- [ ] Terapkan full-text search title/body dengan konfigurasi `simple`, pencocokan semua kata, dan SQL berparameter.
- [ ] Terapkan pencocokan penuh nama author tanpa membedakan kapitalisasi; dua author bernama sama tetap dapat menghasilkan artikel masing-masing.
- [ ] Gabungkan search dan author filter menggunakan AND, termasuk saat mengambil halaman lanjutan.
- [ ] Terapkan limit dan cursor berdasarkan pasangan timestamp/ID; cursor tidak valid ditolak. Klien mengulang dari halaman pertama saat mengganti filter.
- [ ] Tambahkan migrations index GIN untuk search, `lower(authors.name)`, urutan artikel, dan `(author_id, created_at, id)` sesuai PRD.
- [ ] Uji title-only/body-only match, beberapa kata lintas field, non-match substring, variasi kapitalisasi, filter gabungan, timestamp sama, dan halaman terakhir.

Selesai jika: hasil sesuai filter, tetap newest-first, dan pagination tidak mengulang atau melewatkan artikel pada dataset tetap. Perilaku live feed saat ada insert baru terdokumentasi.

## T07 - Verifikasi Concurrency dan Query

Dependensi: T05 dan T06. Acuan: PRD 8 dan 9.

- [ ] Pastikan controller/usecase/repository tidak menyimpan data request dalam mutable state bersama.
- [ ] Tetapkan batas connection pool dan timeout sebagai konfigurasi; dokumentasikan nilai awal sebagai asumsi, lalu sesuaikan berdasarkan pengukuran.
- [ ] Verifikasi cancellation/timeout sampai SQL, pelepasan koneksi, dan penanganan error ketika operasi database gagal atau menunggu pool.
- [ ] Jalankan create dan list/search secara bersamaan; cocokkan ID dari respons create sukses dengan data tersimpan dan periksa tidak ada partial record.
- [ ] Periksa query plan pada data representatif untuk list, search, author filter, dan kombinasi filter; evaluasi penggunaan index.
- [ ] Catat ukuran dataset, concurrency, durasi, lingkungan, latency, throughput, serta error selama load check sederhana. Tool tidak wajib k6; tidak ada target performa numerik dari assessment.

Selesai jika: hasil pengujian concurrency dan query terdokumentasi beserta keterbatasannya. Atomic insert tidak diklaim sebagai jaminan idempotency untuk POST yang diulang klien.

## T08 - Verifikasi Akhir

Dependensi: T04-T07. Acuan: PRD 9 dan 11.

- [ ] Jalankan unit/integration test menggunakan database test terpisah; lengkapi kasus gagal yang belum tercakup pada T05-T07.
- [ ] Jalankan `go test ./...`, `go vet ./...`, dan `go test -race ./...` pada lingkungan yang mendukung race detector; pastikan tes integrasi benar-benar dijalankan.
- [ ] Ulangi setup Docker dari lingkungan bersih: konfigurasi, migrations, seed, health, create, list, search, dan filter.
- [ ] Buat artikel, recreate container dengan volume tetap dipertahankan, lalu pastikan artikel masih tersedia.

Selesai jika: tes wajib lulus dan setup Docker dapat direproduksi tanpa mengandalkan state lokal tersembunyi.

## T09 - Dokumentasi dan Review Submission

Dependensi: T08. Acuan: PRD 3, 5, dan 11.

- [ ] Lengkapi README: prerequisites, Docker setup, environment, migrations/seed, contoh author ID, contoh request, serta cara menjalankan tes.
- [ ] Catat keputusan API, batas pagination, konfigurasi concurrency, hasil load/query check, dan keterbatasan yang ditemukan.
- [ ] Jelaskan alur layer, strategi index, dan alasan pemilihan solusi secara singkat.
- [ ] Ungkapkan bagian yang dibantu AI, termasuk PRD, task breakdown, health endpoint/test, dan implementasi berikutnya jika menggunakan AI.
- [ ] Cocokkan hasil dengan seluruh acceptance criteria PRD dan pastikan kandidat dapat menjelaskan kode serta tradeoff-nya.

Selesai jika: reviewer bisa menjalankan, menguji, dan memahami proyek dari README; seluruh kebutuhan assessment dan keputusan proyek tercakup.

## T10 - Opsional: Load Test k6

Dependensi: T08. Bukan syarat selesai assessment atau pengganti tes correctness.

- [ ] Jika masih ada waktu, buat satu script k6 lokal berisi campuran create, list, search, dan author filter dengan data yang representatif.
- [ ] Periksa status/respons dan catat request rate, latency p95, error rate, jumlah data, beban, durasi, serta spesifikasi lingkungan.
- [ ] Tambahkan perintah dan hasilnya ke README; hasil dapat melengkapi pengukuran T07 tanpa menambahkan stack monitoring.

Selesai jika dipilih: script dapat dijalankan ulang dan hasilnya membantu menjelaskan performa pada workload yang diuji.

Tetap di luar scope: update/delete/detail artikel, author-management API, autentikasi, frontend, kategori, komentar, cache, search cluster terpisah, queue, microservices, dan distributed lock.
