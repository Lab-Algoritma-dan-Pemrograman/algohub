---
title: Database & Authentication Architecture
description: Supabase PostgreSQL pooler configuration on port 6543, multi-token JWT authentication, application-level RLS enforcement, and rate-limiting.
---

# Database & Authentication Architecture

AlgoHub operates a centralized **Supabase PostgreSQL** database engine serving four distinct frontend clients, governed by a unified multi-token JWT authentication schema and application-level Row Level Security (RLS) enforcement inside the Go backend.

---

## 1. Supabase PostgreSQL Core Infrastructure

All subsystems (Praktikum, SIAKAD, and E-Learning) connect to a shared PostgreSQL database cluster (`oidwvfcxlxfyojkyurmm`).

```
                    +------------------------------------+
                    |        Backend Go Serverless       |
                    |          (api.algohub.web.id)      |
                    +-----------------+------------------+
                                      |
                                      | Transaction Pooler
                                      | Port 6543 (pgx driver)
                                      v
                    +------------------------------------+
                    |      Supabase Supavisor Pooler     |
                    +-----------------+------------------+
                                      |
                                      | Internal Unix Socket
                                      v
                    +------------------------------------+
                    |     PostgreSQL Core Database       |
                    |  (Schemas: public, auth, storage)  |
                    +------------------------------------+
```

### Connection Pooler Configuration (Port 6543)
Because the Go backend runs on Vercel Fluid Compute serverless infrastructure with elastic horizontal autoscaling, connections are not established directly against PostgreSQL port 5432 (which imposes rigid connection caps). Instead, all traffic routes through the **Supavisor Transaction Pooler** on port `6543`.

GORM and `pgx` driver parameters are tuned specifically:
- **`PreferSimpleProtocol: true`**: **Strictly mandatory**. In transaction pooling mode, distinct physical database connections are recycled between consecutive statements. PostgreSQL binary prepared statements trigger naming collisions under this model. Simple protocol mode transmits queries as parameterized plain SQL strings.
- **Serverless Pool Dimensions**:
  - `MaxOpenConns`: 5 (strictly limits concurrent database sockets per function instance)
  - `MaxIdleConns`: 2
  - `ConnMaxIdleTime`: 1 minute
  - `ConnMaxLifetime`: 5 minutes
- **Zero Runtime AutoMigrate**: Schema migrations and automatic seed routines are completely disabled in production builds to prevent concurrent table locks during academic exams.

---

## 2. Multi-Token JWT Authentication System

Each AlgoHub frontend utilizes session credentials optimized for its specific operational profile:

| Module | Session Vehicle | Governing Secret Key | Primary JWT Claims |
|---|---|---|---|
| **Praktikum** | `Authorization: Bearer <token>` in `localStorage` | `JWT_SECRET` (HS256) | `{ uid, nim, role, exp }` |
| **SIAKAD** | HTTP-Only Cookie `siakad_token` on `.algohub.web.id` + Bearer | `SIAKAD_JWT_SECRET` (HS256) | `{ username, full_name, role, exp }` |
| **E-Learning** | `Authorization: Bearer <token>` | `ELEARNING_SUPABASE_JWT_SECRET` (HS256) | `{ nim, sub, role: "authenticated", user_role }` |

### Backend Authentication Modes (`AUTH_MODE`)
The Go backend supports three runtime verification modes controlled by the `AUTH_MODE` environment variable:
1. `legacy`: Validates native Go HS256 tokens only.
2. `supabase`: Validates Supabase RS256/HS256 tokens via JWKS keysets.
3. `dual`: **Active Production Mode**. The backend inspects inbound signatures and routes verification to the appropriate token parser (Praktikum tokens, SIAKAD cookies, or E-Learning re-signed Supabase tokens).

### Centralized Role Normalization (`pkg/jwt/roles.go`)
To resolve historical terminology divergences across services, `NormalizeRole()` standardizes claims into four canonical roles:
- `mahasiswa`, `user`, `student` $\longrightarrow$ `praktikan`
- `kordas`, `coordinator`, `super_admin` $\longrightarrow$ `koordinator`
- `admin`, `laboran`, `assistant` $\longrightarrow$ `asisten`
- `penyewa` $\longrightarrow$ `penyewa`

### Password Hash Backwards Compatibility (Bcrypt & Firebase Scrypt)
Legacy accounts created during the Firebase era authenticate transparently using `pkg/hash/scrypt.go` (`FB_SCRYPT_*`). When users update credentials, the system automatically migrates hashes to standard **Bcrypt**.

---

## 3. Application-Level RLS Enforcement in Go Usecases

The Go backend connects to PostgreSQL as the `postgres` superuser role. Under standard PostgreSQL semantics, `postgres` **automatically bypasses all table-level RLS policies**.

Consequently, **all 56 active PostgreSQL RLS security policies are replicated and enforced within the Go usecase layer (`usecase/elearning`)**:

```
Client Request -> Router -> Middleware Auth (Extract NIM & Role) -> Usecase Check:
  |
  +-- Is student attempting to read another peer's progress?
  |     -> Yes: Return filtered data / empty array (RLS SELECT semantics)
  |
  +-- Is student attempting to INSERT into curriculum tables?
  |     -> Yes: Reject with HTTP 403 "violates row-level security"
  |
  +-- Is student attempting to mutate another user's row?
  |     -> Yes: Update 0 rows, return HTTP 200 (RLS UPDATE semantics)
  |
  v Validation Cleared
Call Repository -> Execute Parameterized SQL against PostgreSQL
```

### Protecting Solutions via `student_lessons`
In E-Learning, the `lessons` table stores exam answer keys and coding solutions. Students are never permitted to read raw `lessons` records directly; usecase logic redirects student queries to the `student_lessons` database view, which strips `correct_answer`, `solution`, and hidden test case columns.

---

## 4. Rate-Limiting Protection Architecture

To mitigate brute-force authentication attacks and Denial of Service (DoS):
- **Trusted Client IP Extraction**: `internal/middleware/client_ip.go` inspects the initial entry of the `X-Forwarded-For` or `X-Real-IP` header set by Vercel Edge infrastructure.
- **Rate Limit Tiers**:
  - `POST /api/praktikum/auth/login`: 10 attempts/minute/IP.
  - `POST /api/praktikum/auth/cek-nim`: 20 requests/minute/IP.
  - `POST /api/praktikum/run` (Glot runner fallback): 30 requests/minute/user ID.
  - `POST /api/siakad/auth/login`: 5 attempts/minute/IP.
- **429 Throttling Responses**: Requests exceeding thresholds receive an HTTP `429 Too Many Requests` status code with a `Retry-After: <seconds>` header. Standard CORS headers are preserved on 429 payloads to prevent browser-side CORS masking.
