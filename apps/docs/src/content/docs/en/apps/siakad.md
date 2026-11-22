---
title: SIAKAD Application
description: Laboratory academic operations, dynamic QR attendance, assistant shift scheduling, push/WhatsApp alerts, and the Postgres RPC Gateway architecture.
---

# SIAKAD Application (`apps/siakad`)

**AlgoHub SIAKAD** (`siakad.algohub.web.id`) is the Academic Information System orchestrating laboratory administrative processes: student attendance, teaching assistant duty scheduling, laboratory equipment rentals, and assistant shift-swap workflows.

---

## 1. Core Modules & Routing

SIAKAD includes 21 dedicated pages protected by Role-Based Access Control (RBAC):

| Module | Route | Access Level | Description |
|---|---|---|---|
| **Dashboard** | `/beranda` | All Authenticated | Lab announcements, upcoming schedule, and personal attendance status |
| **Attendance Scanner** | `/absensi` | All Authenticated | Dynamic QR code camera scanner for check-in and check-out |
| **Validate Attendance** | `/validasi-absensi` | Assistant, Coordinator | Manual or batch verification of student lab attendance |
| **QR Code Generator** | `/buat-qr` | Assistant, Coordinator | Rotating Dynamic QR Code token generator displayed during lab sessions |
| **Duty Schedule** | `/jadwal-jaga` | Assistant, Coordinator | Assistant lab duty calendar and shift-swap request interface |
| **My Schedule** | `/jadwal-saya` | All Authenticated | Individual student practicum schedules or assigned assistant shifts |
| **Lab Inventory** | `/inventaris` | Assistant, Coordinator | Equipment and hardware catalog with real-time availability tracking |
| **Equipment Rental** | `/sewa-barang`, `/manajemen-sewa` | All Authenticated / Staff | Hardware loan submissions and staff approval processing |
| **Financial Ledger** | `/laporan-keuangan` | Assistant, Coordinator | Cash bookkeeping, balance sheets, and Microsoft Excel (`.xlsx`) export |
| **Audit Logs** | `/audit-log` | Assistant, Coordinator | Comprehensive audit trail of privileged administrative operations |
| **System Settings** | `/pengaturan` | Coordinator | Global platform configuration, WhatsApp templates, and system toggles |

---

## 2. Dynamic QR Code Attendance

The attendance subsystem mitigates attendance fraud using rotating dynamic tokens:
1. **Token Generator (`/buat-qr`)**: The presiding assistant activates attendance on a lab projection display. The system generates encrypted QR codes refreshed at rapid intervals to prevent unauthorized remote scanning.
2. **Webcam Scanner (`/absensi`)**: Students scan the projection using their mobile or laptop camera powered by the `html5-qrcode` library.
3. **Backend Validation**: The RPC Gateway confirms timestamp validity and student identity via the database function `record_attendance_secure`.

---

## 3. Assistant Duty Scheduling & Shift Swapping

When teaching assistants encounter scheduling conflicts with academic duties:
1. **Swap Submission**: The requesting assistant selects an assigned shift and nominates a peer replacement on `/jadwal-jaga`.
2. **Instant Dispatch**: The platform dispatches an immediate Web Push notification and automated WhatsApp message to the candidate replacement.
3. **Automated Handshake**: Upon acceptance by the peer assistant, the database updates shift allocations without requiring coordinator intervention.

---

## 4. Multi-Channel Notification Infrastructure

SIAKAD integrates two real-time notification channels:

### Web Push Notifications (VAPID)
- Operates via a dedicated service worker (`/service-worker.js`).
- Handled in the Go backend using the `SherClockHolmes/webpush-go` library.
- Transmits immediate alerts regarding duty shifts, swap updates, or report status directly to desktop and mobile browsers.

### WhatsApp Notification Gateway
- Dispatches automated attendance verifications, duty shift reminders, and swap requests to verified mobile numbers.
- Provider endpoints and tokens are configurable in the `system_settings` table.
- Standardizes international phone numbering formats (`+62` / `08xx`).

---

## 5. Postgres RPC Gateway Architecture

The SIAKAD frontend does not perform raw SQL queries or direct table mutations against Supabase. All data interactions are routed through the Go **RPC Gateway** (`POST /api/siakad/rpc?fn=<function_name>`).

```
  +-----------------------------------------------------------+
  | Frontend SIAKAD (@algohub/api-client/siakad)              |
  | api.rpc('get_attendance_records_secure', { session_id })  |
  +-----------------------------+-----------------------------+
                                | HTTP POST (Cookie / Bearer)
                                v
  +-----------------------------------------------------------+
  | Backend Go Gin: /api/siakad/rpc?fn=...                    |
  | 1. Role Normalization (ROLE_ALIASES)                      |
  | 2. Whitelist Permission Validation (ROLE_PERMISSIONS)     |
  | 3. Caller Identity Injection (p_caller_id from JWT)       |
  | 4. Sensitive Credential Scrubbing (stripCredentials)      |
  +-----------------------------+-----------------------------+
                                | SQL Parameterized Query
                                v
  +-----------------------------------------------------------+
  | PostgreSQL Database:                                      |
  | SELECT * FROM public.get_attendance_records_secure(...)   |
  +-----------------------------------------------------------+
```

### Gateway Security Benefits:
1. **Strict Function Whitelisting**: Only vetted Postgres stored procedures can be invoked. Arbitrary calls return `403 Forbidden`.
2. **Tamper-Proof Parameter Injection**: Identifiers such as `p_caller_id`, `p_student_id`, or `p_user_id` are derived from the cryptographically verified JWT rather than client-supplied inputs, eliminating impersonation attacks.
3. **`stripCredentials` Filtering**: Output records are systematically sanitized to prevent accidental leakage of password hashes or encryption salts.
