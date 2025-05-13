# Tugas Besar 2 IF2211 Strategi Algoritma

## Pencarian Recipe pada Little Alchemy 2 dengan Algoritma BFS dan DFS

## 📌 Deskripsi Singkat

Program ini menyelesaikan permasalahan pencarian recipe pada permainan **Little Alchemy 2** menggunakan algoritma **Breadth-First Search (BFS)** dan **Depth-First Search (DFS)**. Pemain dapat mencari recipe untuk membentuk elemen tertentu dari elemen dasar yang tersedia, yaitu **water, fire, earth, air**. Program juga mendukung pencarian banyak recipe (multiple recipes) dengan optimasi multithreading.

Aplikasi berbasis web ini dibangun dengan menggunakan **React.js (Frontend)** dan **Golang (Backend)**, serta memvisualisasikan recipe yang ditemukan dalam bentuk tree. Selain itu, pengguna dapat memilih algoritma pencarian (BFS atau DFS) dan mode pencarian (Single Recipe atau Multiple Recipes) secara langsung melalui antarmuka aplikasi.

---

## 🗂️ Struktur Direktori

```
Tubes2_alchendol/
├── doc
│   └── alchendol.pdf
├── src
│   ├── backend
│   │   ├── api
│   │   │   └── handler.go
│   │   ├── data
│   │   │   └── elements.json
│   │   ├── models
│   │   │   └── models.go
│   │   ├── scrape
│   │   │   └── scraper.go
│   │   ├── search
│   │   │   ├── bidirectional.go
│   │   │   ├── bidirectional_multiple.go
│   │   │   ├── bfs.go
│   │   │   ├── bfs_multiple.go
│   │   │   ├── dfs.go
│   │   │   └── dfs_multiple.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── main.go
│   └── frontend
│       ├── app
│       │   ├── components
│       │   │   ├── BorderBox.js
│       │   │   ├── Button.js
│       │   │   ├── Card.js
│       │   │   ├── Navbar.js
│       │   │   ├── QuantityInput.js
│       │   │   ├── SearchBar.js
│       │   │   ├── Toggle.js
│       │   │   ├── TreeDiagram.js
│       │   │   └── Typography.js
│       │   ├── creator
│       │   │   └── page.js
│       │   ├── greets
│       │   │   └── page.js
│       │   ├── howtoplay
│       │   │   └── page.js
│       │   ├── magicpath
│       │   │   └── page.js
│       │   ├── multiplerecipes
│       │   │   └── page.js
│       │   ├── overview
│       │   │   └── page.js
│       │   ├── result
│       │   │   └── page.js
│       │   ├── search
│       │   │   └── [element]
│       │   │       └── page.js
│       │   ├── searching
│       │   │   └── page.js
│       │   ├── shortestrecipe
│       │   │   └── page.js
│       │   ├── layout.js
│       │   ├── page.js
│       │   └── globals.css
│       ├── data
│       │   └── data.js
│       ├── lib
│       │   └── utils.js
│       ├── public
│       │   ├── icons
│       │   └── img
│       ├── Dockerfile
│       ├── node_modules/
│       ├── .gitignore
│       ├── eslint.config.mjs
│       ├── jsconfig.json
│       ├── next.config.mjs
│       ├── package-lock.json
│       ├── package.json
│       └── postcss.config.mjs
├── docker-compose.yml
└── README.md
```

---

## ⚙️ Cara Menjalankan Program

### 1️⃣ Clone repository:

```bash
git clone https://github.com/adndax/Tubes2_alchendol.git
```

### 2️⃣ Masuk ke direktori frontend:

```bash
cd Tubes2_alchendol/src/frontend
```

### 3️⃣ Install dependencies:

```bash
npm install
```

### 4️⃣ Jalankan aplikasi frontend:

```bash
npm run dev
```

### 5️⃣ Buka tab terminal baru, lalu masuk ke direktori backend:

```bash
cd ../backend
```

### 6️⃣ Jalankan aplikasi backend:

```bash
go run main.go
```

### 7️⃣ Akses aplikasi di browser:

Buka [http://localhost:5173](http://localhost:5173) untuk melihat antarmuka aplikasi.

---

## 📄 Format Input

* Pilihan algoritma (BFS, DFS, dan Bidirectional)
* Elemen yang ingin dicari
* Mode pencarian: Single Recipe atau Multiple Recipes

---

## 🧾 Format Output

* Visualisasi tree recipe yang ditemukan
* Waktu pencarian dan jumlah node yang dikunjungi

---

## 📈 Fitur Tambahan

* **Multithreading** pada mode pencarian Multiple Recipes
* **Visualisasi tree** yang interaktif
* **Live Update** untuk menampilkan proses pencarian secara real-time (bonus)
* **Bidirectional Search** sebagai opsi algoritma tambahan (bonus)
* **Docker Support** untuk containerization (bonus)

---

## 👤 Author

| No | Nama              | NIM      |
| -- | ----------------- | -------- |
| 1  | Muhammad Alfansya | 13523005 |
| 2  | M Hazim R Prajoda | 13523009 |
| 3  | Adinda Putri | 13523071 |
