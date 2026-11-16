# AlgoHub

Monorepo Lab Algoritma dan Pemrograman: frontend SvelteKit (SPA) per subdomain dan satu API Go di Vercel.

| Folder           | Isi             | Domain                     |
| ---------------- | --------------- | -------------------------- |
| `apps/portal`    | landing page    | `algohub.web.id`           |
| `apps/siakad`    | SIAKAD          | `siakad.algohub.web.id`    |
| `apps/elearning` | E-Learning      | `elearning.algohub.web.id` |
| `apps/praktikum` | Praktikum       | `praktikum.algohub.web.id` |
| `backend`        | API Go          | `api.algohub.web.id`       |
| `packages/*`     | kode TS bersama | —                          |

## Syarat

- Bun 1.3.14 (`packageManager` di `package.json`). **Jangan pakai npm, npx, yarn, atau pnpm.**
- Go: versi diatur `toolchain` di `backend/go.mod`, diunduh otomatis (`GOTOOLCHAIN=auto`).
- golangci-lint v2.

## Perintah

```sh
bun install
bun run dev:api      # API lokal di http://127.0.0.1:8080
bun run lint
bun run check
bun run test
bun run build
```

Backend (dari `backend/`):

```sh
go vet ./...
golangci-lint run ./...
go test ./...
```

## Aturan

Setiap perubahan di `apps/`, `backend/`, atau `packages/` wajib disertai entri di [CHANGELOG.md](CHANGELOG.md). Standar kode, aturan changelog, sub-agent, dan urutan fase ada di `E:\rekapProject\project_ap\_update\2026-11-17-monorepo-sveltekit-go\README.md`.
