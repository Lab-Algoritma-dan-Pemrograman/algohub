---
title: Introduction to AlgoHub
description: Integrated platform for the Algorithm and Programming Laboratory supporting coursework, practicum, and academic administration.
---

# Introduction to AlgoHub

**AlgoHub** is an integrated digital platform engineered to power the end-to-end operations of the **Algorithm & Programming Laboratory**. The platform combines real-time interactive lab practicum sessions, self-paced gamified programming curriculum, digital attendance tracking, and comprehensive academic laboratory resource management into a modern, unified ecosystem.

---

## Vision & Mission

- **Vision**: To be a premier computing laboratory center providing an automated, interactive, and high-efficiency environment for honing students' programming problem-solving skills.
- **Mission**:
  1. Deliver seamless in-browser C and Python execution environments eliminating local machine setup friction.
  2. Accelerate algorithm comprehension through multi-tier gamified progression (skill levels and Bug Hunt debugging challenges).
  3. Streamline laboratory operations with dynamic QR attendance, assistant duty rotation, multi-channel notifications, and data-driven evaluation.
  4. Maintain a robust, cost-effective, and secure software architecture built for zero-downtime reliability.

---

## The AlgoHub Ecosystem

AlgoHub adopts a **Monorepo** architecture separating client-side presentation into 4 distinct Single Page Applications (SvelteKit SPA) and a centralized Go Gin REST API service, powered by a shared Supabase PostgreSQL database.

```
                    +---------------------------+
                    |       Cloudflare DNS      |
                    | (DNS-Only / Gray Cloud)   |
                    +-------------+-------------+
                                  |
                +-----------------+-----------------+
                |                                   |
                v                                   v
   +-------------------------+         +-------------------------+
   |   4 Frontend SPAs       |         |   1 Central Backend     |
   |   (SvelteKit Static)    |         |   (Go Gin Serverless)   |
   |   - Portal              |         |   - api.algohub.web.id  |
   |   - SIAKAD              |         +------------+------------+
   |   - E-Learning          |                      |
   |   - Praktikum           |                      | Transaction Pooler
   +------------+------------+                      | (Port 6543)
                |                                   v
                | CORS Fetch           +-------------------------+
                +--------------------->|   Supabase PostgreSQL   |
                                       |   (Shared Database)     |
                                       +-------------------------+
```

### 1. Portal (`algohub.web.id`)
Public landing page and gateway to the AlgoHub ecosystem. Hosts the laboratory profile, directs users to active subsystems, serves immutable pre-built C WebAssembly compiler binaries (`/compiler/v1/`), and handles 301 permanent redirects from legacy practicum bookmarks.

### 2. SIAKAD (`siakad.algohub.web.id`)
Academic and administrative laboratory information system:
- Rotating Dynamic QR Code attendance system for students and assistants.
- Lab duty schedule management and peer shift-swap approval workflows.
- Lab hardware inventory rental tracking.
- Web Push (VAPID) and WhatsApp automated gateway notifications.
- Secure Postgres RPC Gateway proxying client queries.

### 3. E-Learning (`elearning.algohub.web.id`)
Gamified programming education platform:
- Hierarchical course structure: **Level** &rarr; **Module** &rarr; **Lesson**.
- In-browser compiler runtime (Clang WebAssembly & Pyodide Python).
- Daily study streak retention and active study time accounting (*study seconds*).
- XP mechanics, weekly leaderboards, level thresholds, and *Prestige Reset* badges.
- *Bug Hunt* interactive debugging mini-game.
- Seamless Single Sign-On (SSO) authentication initiated from SIAKAD.

### 4. Praktikum (`praktikum.algohub.web.id`)
The core exam and laboratory session execution platform:
- Timed assessment workflows: Pretest, Posttest, Skill Test, and Semester Exams.
- Autosave of answers and code to PostgreSQL, with resilient `localStorage` drafts that survive refreshes.
- Zero-server-overhead C code compilation directly in web browser workers using WebAssembly.
- Final Comprehensive Report submission via Vercel Blob, with Supabase Storage as fallback.
- Dual grading system: assistant manual scoring alongside AI-assisted rubric evaluation.

### 5. Backend Go API (`api.algohub.web.id`)
Unified RESTful API powering all frontend interfaces:
- Built with Go 1.26 and Gin framework, deployed on Vercel Fluid Compute.
- Implements dual-token JWT verification, rate limiting, and business-layer Row Level Security (RLS) enforcement.

---

## Subdomain Directory

All services run under the parent domain `algohub.web.id`:

| Subdomain | Application / Purpose | Technology Stack | Access Tier |
|---|---|---|---|
| `algohub.web.id` | Portal & Landing Page | SvelteKit (Prerendered HTML) | Public |
| `siakad.algohub.web.id` | Academic Administration | SvelteKit SPA + RPC Gateway | Students, Assistants, Coordinators |
| `elearning.algohub.web.id` | Gamified Coursework | SvelteKit SPA + PWA | Students, Assistants, Coordinators |
| `praktikum.algohub.web.id` | Lab Practicum Sessions | SvelteKit SPA + Web Workers | Students, Assistants, Coordinators |
| `api.algohub.web.id` | Central Backend API | Go Gin (Serverless Function) | Token Auth / Cron Secrets |
| `docs.algohub.web.id` | Official Documentation | Astro Starlight | Public |

---

## User Roles & Access Hierarchy

AlgoHub employs an access control hierarchy:
1. **Student / Practikan (`praktikan`)**: Access assigned lab rooms, submit answers, study gamified modules, track streaks, and scan QR attendance.
2. **Renter (`penyewa`)**: Authorized users borrowing laboratory equipment and hardware.
3. **Teaching Assistant (`asisten`)**: Activate lab rooms, monitor student progress, grade submissions, manage shift duties, and validate attendance.
4. **Coordinator (`koordinator`)**: Full administrative control over curriculum, grading schemas, user roles, and operational schedules.
5. **Superadmin**: Global system configuration and underlying permission management.
