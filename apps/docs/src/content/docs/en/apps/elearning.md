---
title: E-Learning Application
description: Gamified programming learning platform, tiered curriculum, WebAssembly & Pyodide code runners, daily study streaks, and SIAKAD SSO.
---

# E-Learning Application (`apps/elearning`)

**AlgoHub E-Learning** (`elearning.algohub.web.id`) is a self-paced, gamified interactive learning platform engineered to reinforce students' algorithmic problem-solving and programming skills outside scheduled laboratory hours.

---

## 1. Curriculum Structure & Content Formats

Curriculum content follows a three-tiered hierarchy:

$$\text{Level} \longrightarrow \text{Module} \longrightarrow \text{Lesson}$$

1. **Level**: Proficiency milestones (e.g. *Foundational C*, *Branches & Loops*, *Arrays & Strings*, *Functions & Pointers*, *Data Structures*).
2. **Module**: Thematic topic clusters grouped under a specific level.
3. **Lesson**: Individual learning units completed by students.

### Interactive Lesson Modalities
Lessons blend diverse pedagogic formats:
- **Rich Text / Markdown Articles**: Conceptual explanations supplemented with formatted syntax, flow diagrams, and architectural summaries.
- **Multiple Choice Quizzes**: Concept checks with solution hashes verified on the server.
- **Flowchart Puzzles**: Drag-and-drop structural logic puzzles requiring students to model algorithm control flow prior to code implementation.
- **Hands-On Coding Exercises**: Monaco Code Editor integration with automated test suite validation.

---

## 2. In-Browser Code Runner (Clang WASM & Pyodide)

E-Learning executes student code entirely client-side using the `@algohub/runner` package:

```
  +-----------------------------------------------------------+
  | Lesson Page / Playground (Main UI Thread)                 |
  | Monaco Code Editor -> "Run Code" Trigger                  |
  +-----------------------------+-----------------------------+
                                | postMessage()
                                v
  +-----------------------------------------------------------+
  | Isolated Web Worker (@algohub/runner)                     |
  | - C Language      : Clang WebAssembly + POSIX memfs layer |
  | - Python Language : Pyodide (CPython WebAssembly Runtime) |
  +-----------------------------+-----------------------------+
                                | Input/Output via SharedArrayBuffer
                                v
  +-----------------------------------------------------------+
  | xterm Terminal Display / Matplotlib Plot Canvas           |
  +-----------------------------------------------------------+
```

- **C Execution**: Powered by Clang & LLD WebAssembly binaries served from `/compiler/v1/`. Interactive terminal inputs (`scanf`) are synchronized via `SharedArrayBuffer` and `Atomics.wait()`.
- **Python Execution**: Driven by Pyodide WebAssembly. Supports scientific computing packages and renders **Matplotlib** figures directly into canvas image arrays (`images[]`).
- **Sandboxed Fault Isolation**: Runaway execution loops (e.g. `while(1)`) can be killed instantly by terminating the underlying worker without affecting browser responsiveness.

---

## 3. Gamification Mechanics & Study Retention

E-Learning leverages behavioral feedback loops to cultivate continuous learning habits:

### Daily Streaks
- Tracks each student's last active engagement timestamp (`last_active_date`).
- Completing at least one lesson or exercise per calendar day increments the continuous streak counter.
- Missing an entire calendar day resets the active streak to zero.

### Active Study Time Accounting (*Study Seconds*)
- In-browser listeners track verified user engagement during active lesson tabs.
- At 30-second intervals, accumulated study time is committed to the backend via a batched endpoint `PUT /api/elearning/sessions/me` to conserve serverless quota.

### Leaderboards & Weekly Resets
- **Experience Points (XP)**: Earned through lesson completion, quizzes, and code challenges.
- **Account Levels**: Calculated dynamically in the database via `calculate_level(xp)`.
- **Global & Class Leaderboards**: Filterable by class code and major.
- **Weekly Cycles**: Weekly leaderboards reset every **Monday at 00:00 WIB**, creating recurring competitive sprints.

### Voluntary Prestige Reset
Students who achieve the maximum curriculum level can trigger a **Prestige Reset** via the RPC procedure `do_prestige_reset`:
- Resets total XP back to 0.
- Awards an exclusive Prestige Medal badge on the student's public profile.
- Unlocks advanced honor challenges.

---

## 4. Mini-Game: Bug Hunt

**Bug Hunt** is an interactive debugging game presenting students with C or Python snippets containing a single deliberate bug:
- Players must identify and select the erroneous line before the countdown timer expires.
- Regulated by a weekly play allotment to maintain balanced XP economics.
- Play rights and quotas are validated via `GET /api/elearning/game/plays/me`.

---

## 5. Single Sign-On (SSO) from SIAKAD

Students seamlessly navigate from academic administrative portals into E-Learning without separate credentials:

```
  +--------------+            +-------------------+            +-----------------+
  | SIAKAD       |            | Backend Go API    |            | E-Learning      |
  | (siakad.*)   |            | (api.*)           |            | (elearning.*)   |
  +-------+------+            +---------+---------+            +--------+--------+
          |                             |                               |
          | 1. Form POST /api/sso       |                               |
          +---------------------------->|                               |
          |                             | 2. Generate Code AES-256-GCM  |
          |                             |    (TTL <= 60 seconds)        |
          | 3. HTTP 303 Redirect        |                               |
          |<----------------------------+                               |
          |                                                             |
          | 4. Open URL: /sso?code=<one_time_code>                      |
          +------------------------------------------------------------>|
          |                                                             |
          |                             | 5. POST /receive-token {code} |
          |                             |<------------------------------+
          |                             | 6. Elearning JWT Session      |
          |                             +------------------------------>|
          |                                                             |
          |                                                             | 7. Enter Dashboard
```

1. SIAKAD dispatches user authentication parameters via form POST to `/api/elearning/sso`.
2. The Go backend produces a single-use token encrypted using AES-256-GCM (TTL $\le$ 60 seconds).
3. The browser receives an HTTP 303 Redirect pointing to `https://elearning.algohub.web.id/sso?code=...`. Long-lived credentials are never exposed in URL parameters.
4. The client `/sso` route exchanges this temporary ticket via `POST /api/elearning/receive-token` to establish an authenticated session.
