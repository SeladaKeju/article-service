# Kontrak API Artikel

Hasil T01, mengacu pada [PRD](PRD.md) dan [task breakdown](TASK_BREAKDOWN.md). Kontrak ini diterapkan oleh endpoint artikel. Detail format, validasi, dan pagination merupakan keputusan/asumsi proyek, bukan requirement tambahan dari assessment. Penyusunan dokumen dibantu AI.

## 1. Format Umum

- Request create dan seluruh respons artikel/error menggunakan JSON.
- Respons sukses menggunakan envelope `data`. Respons list menambahkan `next_cursor`, tanpa total count.
- Artikel memiliki tepat lima field: `id`, `author_id`, `title`, `body`, dan `created_at`. Nama author tidak disertakan.
- ID artikel dan author adalah string [UUID v4](https://www.rfc-editor.org/rfc/rfc9562.html#name-uuid-version-4), dikirim dalam bentuk lowercase berhipen dan disimpan sebagai `text`, sesuai diagram assessment. UUID input berhipen boleh menggunakan huruf kapital; normalisasikan ke lowercase setelah validasi.
- Server menghasilkan ID artikel dan `created_at`. Timestamp disimpan sebagai waktu UTC dan dikirim dalam [RFC 3339](https://www.rfc-editor.org/rfc/rfc3339.html) dengan tepat enam digit pecahan detik dan akhiran `Z`. Respons dan cursor menggunakan nilai timestamp yang benar-benar tersimpan.
- Author tersedia melalui migration seed `000003`. ID/nama seed pada contoh di bawah tersedia setelah migrations dijalankan.

## 2. Create Article

`POST /articles`, dengan `Content-Type: application/json`.

### Request

```json
{
  "author_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Go Concurrency",
  "body": "Concurrent requests in Go."
}
```

Ketiga field wajib bertipe string dan tidak boleh hilang, `null`, kosong, atau hanya berisi whitespace. Tolak field tambahan, termasuk `id` dan `created_at`, JSON rusak, array, scalar, dan lebih dari satu object JSON dalam satu body.

Trim whitespace di tepi `author_id` sebelum validasi UUID v4 dan pencarian author. Untuk title/body, gunakan trim hanya untuk memeriksa apakah isinya kosong; simpan dan kembalikan isi asli tanpa trimming atau perubahan kapitalisasi.

UUID author yang tidak valid menghasilkan `400 invalid_request`. UUID valid yang tidak merujuk author tersedia menghasilkan `400 author_not_found`. Request yang ditolak tidak membuat artikel atau author.

### Respons Sukses: 201 Created

```json
{
  "data": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "author_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Go Concurrency",
    "body": "Concurrent requests in Go.",
    "created_at": "2026-09-08T12:00:00.123456Z"
  }
}
```

Kirim respons sukses setelah penyimpanan berhasil. POST yang diulang klien merupakan request create baru; kontrak ini tidak menjanjikan idempotency.

## 3. List, Search, dan Author Filter

`GET /articles` mengembalikan `200 OK` dengan artikel terbaru lebih dahulu, menggunakan urutan `created_at DESC, id DESC`.

| Parameter opsional | Kontrak |
| --- | --- |
| `query` | Whole-word search, case-insensitive; semua token harus ditemukan di gabungan title/body |
| `author` | Pencocokan nama author secara penuh, case-insensitive |
| `limit` | Integer 1-100; default 20 |
| `cursor` | Posisi lanjutan; tidak diberikan pada halaman pertama |

Trim whitespace di tepi nilai parameter setelah URL decoding. Nilai kosong atau whitespace-only dianggap tidak diberikan, termasuk `limit` dan `cursor`. Limit non-integer atau di luar rentang menghasilkan `400 invalid_limit`; jangan diam-diam membatasi nilai yang tidak valid.

Search memakai tokenisasi PostgreSQL dengan konfigurasi `simple`: tidak ada pencocokan substring, stemming bahasa, fuzzy search, atau pengurutan relevansi. Semua token query harus muncul, tetapi dapat tersebar di title dan body tanpa mengikuti urutan query. Query nonkosong yang tidak menghasilkan token mengembalikan daftar kosong.

Contoh perilaku:

- `query=go` cocok dengan kata `Go`, tetapi tidak dengan `Golang` atau `Django`.
- `query=go%20concurrency` cocok jika title mengandung `Go` dan body mengandung `concurrency`.
- `author=alice` cocok dengan nama `Alice`, tetapi tidak dengan `Alice Smith`.
- Jika dua author berbeda bernama `Alice`, artikel keduanya dapat muncul. Nama author tidak dianggap unik.
- Jika query dan author diberikan bersama, kedua kondisi harus terpenuhi (AND), termasuk pada halaman berikutnya.

### Contoh Filter Gabungan dan Halaman Pertama

Asumsikan author contoh bernama `Alice` dan ada dua artikel yang cocok dengan filter, dengan timestamp sama. Artikel kedua ditampilkan pada contoh halaman berikutnya.

```http
GET /articles?query=go%20concurrency&author=alice&limit=1
```

Respons `200 OK`:

```json
{
  "data": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "author_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Go Concurrency",
      "body": "Concurrent requests in Go.",
      "created_at": "2026-09-08T12:00:00.123456Z"
    }
  ],
  "next_cursor": "eyJjcmVhdGVkX2F0IjoiMjAyNi0wOS0wOFQxMjowMDowMC4xMjM0NTZaIiwiaWQiOiIzZmE4NWY2NC01NzE3LTQ1NjItYjNmYy0yYzk2M2Y2NmFmYTYifQ"
}
```

## 4. Cursor dan Halaman Berikutnya

Cursor adalah JSON dengan tepat `created_at` dan `id` dari artikel terakhir yang dikembalikan, dienkode sebagai [Base64 URL-safe tanpa padding](https://pkg.go.dev/encoding/base64#RawURLEncoding). Cursor contoh di atas didekode menjadi:

```json
{
  "created_at": "2026-09-08T12:00:00.123456Z",
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
}
```

- Gunakan timestamp persis seperti nilai tersimpan; jangan membulatkannya menjadi detik atau milidetik saat membuat cursor.
- Halaman berikutnya mengambil pasangan `(created_at, id)` yang lebih kecil daripada pasangan cursor, dengan filter dan urutan yang sama.
- Ambil maksimal `limit + 1` hasil. Kembalikan maksimal `limit`; buat `next_cursor` dari artikel terakhir yang dikembalikan hanya jika hasil tambahan tersedia.
- Encoding tidak valid, JSON bukan satu object, field hilang/tambahan, tipe salah, timestamp bukan format UTC enam digit di atas, atau ID bukan UUID v4 menghasilkan `400 invalid_cursor`.
- Cursor berisi posisi, bukan snapshot. Klien mempertahankan filter selama pagination dan memulai dari halaman pertama ketika filter berubah. Filter tidak dimasukkan ke payload cursor.
- Live feed tidak menjanjikan snapshot lintas request. Refresh dari halaman pertama untuk melihat artikel baru; jaminan tidak ada item terulang/terlewat diuji pada dataset tetap.

### Request Halaman Berikutnya

```http
GET /articles?query=go%20concurrency&author=alice&limit=1&cursor=eyJjcmVhdGVkX2F0IjoiMjAyNi0wOS0wOFQxMjowMDowMC4xMjM0NTZaIiwiaWQiOiIzZmE4NWY2NC01NzE3LTQ1NjItYjNmYy0yYzk2M2Y2NmFmYTYifQ
```

Respons `200 OK`, jika artikel berikut adalah hasil terakhir:

```json
{
  "data": [
    {
      "id": "2fa85f64-5717-4562-b3fc-2c963f66afa6",
      "author_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Go Patterns",
      "body": "Concurrency with independent requests.",
      "created_at": "2026-09-08T12:00:00.123456Z"
    }
  ],
  "next_cursor": null
}
```

### Hasil Kosong

Contoh request tanpa kecocokan pada data contoh: `GET /articles?author=Nobody`.

Respons `200 OK`; array selalu `[]`, bukan `null`:

```json
{
  "data": [],
  "next_cursor": null
}
```

## 5. Format Error

Semua error menggunakan envelope `error`, tanpa `data`. Code adalah penanda stabil untuk klien; message merupakan penjelasan singkat. Contoh message tidak mewajibkan wording yang persis sama.

| HTTP | Code | Kondisi |
| --- | --- | --- |
| 400 | `invalid_request` | JSON atau field create tidak valid |
| 400 | `author_not_found` | UUID author valid tetapi tidak tersedia |
| 400 | `invalid_limit` | Limit bukan integer atau di luar 1-100 |
| 400 | `invalid_cursor` | Encoding, struktur, timestamp, atau ID cursor tidak valid |
| 500 | `internal_error` | Kegagalan internal/database; pesan umum tanpa detail SQL |

Contoh `400`, ketika title hanya berisi whitespace:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "title must not be blank"
  }
}
```

Contoh `500`:

```json
{
  "error": {
    "code": "internal_error",
    "message": "An internal error occurred"
  }
}
```

## 6. Skenario Verifikasi untuk Implementasi Berikutnya

| Skenario | Hasil yang diharapkan |
| --- | --- |
| Create valid | 201, envelope data, lima field artikel, ID/timestamp server, data tersimpan |
| Field hilang/null/tipe salah/whitespace-only/tambahan | 400 invalid_request; tidak ada artikel dibuat |
| JSON rusak, array/scalar, atau dua object berurutan | 400 invalid_request |
| Author ID dengan whitespace tepi atau kapitalisasi UUID berbeda | Trim/normalisasi ID; author yang sama ditemukan |
| Title/body valid dengan whitespace tepi | Isi asli tetap tersimpan dan dikembalikan |
| UUID author salah / UUID valid tetapi tidak dikenal | 400 invalid_request / 400 author_not_found |
| Kata hanya ada di title / hanya di body | Artikel ditemukan untuk kedua kasus |
| Dua kata tersebar di title/body / hanya satu kata ditemukan | Cocok / tidak cocok |
| query=go terhadap Go, Golang, dan Django | Hanya artikel dengan token Go yang cocok |
| Query kosong / query nonkosong tanpa token | Tanpa filter query / daftar kosong |
| author=alice terhadap Alice dan Alice Smith | Hanya nama penuh Alice yang cocok, termasuk beberapa author dengan nama sama |
| Query dan author gabungan | Kedua kondisi terpenuhi pada semua halaman |
| Limit hilang/kosong, 1, dan 100 | Default 20, batas bawah diterima, batas atas diterima |
| Limit 0, -1, 101, 1.5, atau abc | 400 invalid_limit |
| Cursor round-trip | Timestamp dan ID hasil decode sama dengan artikel terakhir yang dikembalikan |
| Cursor rusak atau field/timestamp/ID tidak valid | 400 invalid_cursor |
| Dua artikel bertimestamp sama | ID descending menentukan urutan; halaman berikutnya tidak mengulang artikel pertama |
| Hasil tepat limit dan tidak ada sisanya / ada satu hasil tambahan | next_cursor null / cursor non-null |
| Tidak ada hasil | 200 dengan data [] dan next_cursor null |
| Create bersamaan dengan pagination | Tidak mengklaim snapshot; refresh memuat artikel baru yang sudah committed |
| Kegagalan database/internal | 500 internal_error tanpa kebocoran detail SQL |

Verifikasi T01 memeriksa validitas contoh JSON, decode/encode cursor, serta konsistensi dokumen. Tes endpoint executable dan pembuktian terhadap PostgreSQL merupakan pekerjaan task implementasi berikutnya.
