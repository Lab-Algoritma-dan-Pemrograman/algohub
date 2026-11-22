---
title: Praktikum Application
description: Core programming examination platform, timed assessment rooms, autosave draft persistence, WebAssembly compiler runner, and final report management.
---

# Praktikum Application (`apps/praktikum`)

**AlgoHub Praktikum** (`praktikum.algohub.web.id`) is the core examination and practical assessment engine supporting all coursework operations in the Algorithm and Programming Laboratory. It powers computer-based examinations, automated code evaluation, final project report archiving, and academic grading gradebooks.

---

## 1. Laboratory Practicum Session Lifecycle

Each formal lab meeting is partitioned into standardized assessment stages:

```
  +------------------+     +-------------------+     +---------------------+
  |     PRETEST      |     |     POSTTEST      |     |     SKILL TEST      |
  | (10-15 Minutes)  |---->|  (15-20 Minutes)  |---->|   (45-60 Minutes)   |
  | Concept Checking |     | Lab Theory Review |     | Live Code Challenge |
  +------------------+     +-------------------+     +---------------------+
                                                                |
                                                                v
                                                     +---------------------+
                                                     |    SEMESTER EXAM    |
                                                     |    (Midterm / Final)|
                                                     | Full Practical Exam |
                                                     +---------------------+
```

### Operational Workflow:
1. **Schedule Verification**: Students check designated meeting times on `/info/jadwal`.
2. **Room Activation**: Teaching assistants unlock room sessions via `/admin/aktivasi` and publish access tokens if required.
3. **Assessment Workspace (`/praktikum/ruang`)**: Students enter the digital workspace containing essay prompts, multiple-choice items, and programming challenges accompanied by an active countdown timer synchronized to the server clock.
4. **Interactive Compilation**: Students author solutions in Monaco Editor and test algorithms against diverse input test cases.
5. **Final Submission**: Students finalize their attempt. Upon timer expiration, the platform performs an automated final submission of all cached responses.

---

## 2. Real-Time Autosave & Data Integrity

To safeguard student progress against hardware faults or accidental tab closures:

### Dual-Layer Storage Strategy
- **Client Storage / IndexedDB (`algohub-praktikum-<id>`)**: Keystrokes and form updates are instantly saved locally in the browser's IndexedDB engine.
- **Debounced Remote Autosave (`POST /api/praktikum/autosave`)**: Edits are periodically synced to the Go backend using debounce intervals, preventing serverless invocation floods while guaranteeing persistence.

### Graceful Auto-Submit Mechanics
- When the client-side countdown reaches `00:00:00`, the editor locks down immediately and transmits final drafts to `/api/praktikum/submit`.
- An automated backend background worker (`/api/cron/praktikum/auto-submit`) periodically audits overdue sessions and freezes unsubmitted attempts past the grace threshold.

---

## 3. In-Browser C WebAssembly Runner

Unlike conventional online judges that execute submissions in heavy server sandboxes, AlgoHub performs all compilation and execution **entirely inside the student's browser**:

```
  +-----------------------------------------------------------+
  | Practicum Room Origin: praktikum.algohub.web.id           |
  | HTTP Headers: Cross-Origin-Opener-Policy: same-origin     |
  | HTTP Headers: Cross-Origin-Embedder-Policy: require-corp  |
  +-----------------------------+-----------------------------+
                                | Establishes crossOriginIsolated
                                v
  +-----------------------------------------------------------+
  | C Compiler Web Worker (@algohub/runner)                   |
  | 1. Fetches Clang WASM from algohub.web.id/compiler/v1/    |
  | 2. Compiles student source code -> output.wasm            |
  | 3. Runs output.wasm within virtual POSIX memory sandbox   |
  | 4. Intercepts scanf() via SharedArrayBuffer & Atomics     |
  +-----------------------------+-----------------------------+
                                | Emits stdout / stderr streams
                                v
  +-----------------------------------------------------------+
  | Interactive Web Terminal (xterm.js)                       |
  +-----------------------------------------------------------+
```

### Architectural Advantages:
- **Zero Server Compute Overhead**: Hundreds of concurrent students compiling programs simultaneously incur zero CPU or memory consumption on backend servers.
- **Fault Containment**: Malicious or runaway constructs (such as memory leaks or fork bombs) remain confined strictly to the client's WebAssembly sandbox.
- **Sub-Second Feedback**: Compilation errors and diagnostic messages render instantaneously without network hops.

---

## 4. Final Comprehensive Report Submissions

At semester conclusion, students submit their final group project documentation:
- **Direct Cloud Upload**: Students obtain presigned S3/Supabase upload URLs via `POST /api/praktikum/laporan-besar/upload-url` and transmit archives directly into encrypted storage buckets.
- **Assistant Verification**: Assistants inspect submissions on `/admin/laporan-besar`, provide revision remarks, and assign official completion status.
- **Batch Export**: Course coordinators can generate and download comprehensive ZIP archives containing all submissions for an entire academic cohort via `POST /api/praktikum/admin/laporan-besar/:id/download-batch`.

---

## 5. Grading Engine & Data Virtualization

- **Standardized Rubrics**: Teaching assistants grade essay and programming solutions against structured scoring rubrics.
- **AI-Assisted Grading**: Integrated AI evaluators (Ollama or Gemini) analyze logic structures and check for plagiarism markers.
- **Virtualized Data Tables**: Gradebook interfaces on `/admin/rekap-nilai` utilize `@tanstack/svelte-virtual` to smoothly render rosters exceeding 1,200 rows with 60 FPS scrolling and minimal DOM nodes.
- **RFC 4180 CSV Export**: Final grades export into RFC 4180-compliant CSV files for submission to university grading portals.
