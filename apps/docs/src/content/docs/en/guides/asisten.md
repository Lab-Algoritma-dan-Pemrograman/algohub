---
title: Assistant & Coordinator Guide
description: Standard operating procedures for laboratory teaching assistants and coordinators covering session activation, exam monitoring, AI grading, and shift duty management.
---

# Assistant & Coordinator Guide

This guide establishes the standard operational procedures for **Teaching Assistants (TAs)** and **Assistant Coordinators (*Kordas*)** to activate lab rooms, monitor live examinations, conduct rubric evaluations, and oversee shift scheduling in AlgoHub.

---

## 1. Practicum Session Activation & Exam Tokens

Teaching assistants assigned to a laboratory shift are responsible for provisioning and supervising student room access:

### Procedure to Activate a Session:
1. Navigate to the admin interface at `https://praktikum.algohub.web.id/praktikum/admin/aktivasi`.
2. Click **Activate New Session**.
3. Configure session parameters:
   - **Class & Module**: Choose the participating academic cohort and syllabus module.
   - **Session Type**: Select the assessment modality (`Pretest`, `Posttest`, `Skill Test`, or `Exam`).
   - **Duration**: Allocate the countdown window in minutes (e.g. Pretest 15 mins, Skill Challenge 60 mins).
4. **Generate Security Token**:
   - Click **Generate Token** to produce an ephemeral 6-character alphanumeric room key.
   - Project this token onto the central laboratory display screen for physical attendees.
5. Click **Open Session**. The session status shifts to Active, permitting students to join.

### Managing Make-Up Assessments (*Susulan*)
For students with excused absences (medical certificates or academic duty):
1. Open the session detail record at `/admin/aktivasi/:id`.
2. Navigate to the **Make-Up Participants** tab.
3. Locate the student by NIM or name, then select **Grant Make-Up Access**.
4. The system unlocks assessment access for the student with a full time allowance, independent of the closed regular cohort.
5. Upon completion, revoke access by clicking **Remove Make-Up Access**.

### Cohort & Individual Access Controls
- **Cohort Lock/Unlock**: If all students finish early, assistants can conclude the session by clicking **Close Session**. This triggers backend routines to auto-submit all pending drafts.
- **Granular Student Controls**: Assistants can selectively grant time extensions or freeze access for individual students encountering hardware malfunctions in the lab.

---

## 2. Real-Time Student Session Monitoring

During live assessments, assistants supervise workspace integrity via the monitoring console:
1. Open **Session Monitoring** at `/admin/monitoring/sesi`.
2. The grid displays live participant telemetry:
   - **Heartbeat Status**: Green indicators verify that the student's browser is actively dispatching keep-alive heartbeats.
   - **Active Question**: Identifies the specific prompt number currently in focus on the student's screen.
   - **Persistence Timestamp**: Displays the most recent timestamp when code was synced to the cloud.
   - **Submission State**: Color-coded badges indicate final locked submissions.
3. **Anomaly Intervention**: If a participant drops offline for more than 2 minutes, the presiding assistant can immediately investigate physical workstation connectivity.

---

## 3. Grading & Academic Evaluation

AlgoHub unifies human evaluation with AI-assisted code analysis tools:

### Essay & Programming Assessment
1. Open **Grading** at `/admin/penilaian/rekap`.
2. Select the completed cohort and session.
3. Click a student record to open the dual-pane comparison workspace:
   - The left pane renders the prompt rubric and reference test cases.
   - The right pane displays the student's code or written response.
4. Input numerical scores (0–100 scale) and attach qualitative feedback notes.

### AI-Assisted Code Grading
To expedite preliminary assessments across hundreds of code submissions:
1. Click **AI-Grade** on the student's submission sheet.
2. The AlgoHub Go backend analyzes source code against test rubrics using integrated language models (local Ollama or Google Gemini).
3. The evaluator outputs a recommended numerical score, runtime complexity analysis, and syntax flaw breakdown.
4. **Human Verification**: Assistants inspect the AI diagnostic, adjust points based on objective judgment, and confirm final grades into the database.

### In-Lab Participation Scoring
Assistants track student engagement, debugging inquiry, and lab discipline through the `/admin/keaktifan` portal.

### Final Report Review & Batch Archiving
1. Open **Laporan Besar** at `/admin/laporan-besar`.
2. Choose the assigned class section.
3. Click **Preview** to review PDF documentation directly in-browser without local file downloads.
4. Set status:
   - **Mark**: Report meets standards; input final grade.
   - **Override**: Request revisions and re-open upload permissions for the student group.
5. **Batch Archive**: For academic audit compliance, click **Download All (ZIP)** to bundle all cohort submissions into a single compressed archive.

---

## 4. Duty Scheduling & Shift Swapping in SIAKAD

### Inspecting Duty Calendars
1. Open `https://siakad.algohub.web.id/jadwal-jaga`.
2. The calendar visualizes your assigned lab supervision shifts across the academic semester.

### Peer Shift Swapping Workflow
In the event of academic conflicts or emergency leave:
1. Locate the duty card on your calendar and click **Request Swap**.
2. Select an agreed peer assistant from the roster.
3. Provide a brief justification note.
4. Click **Submit Request**. The platform immediately sends automated WhatsApp and Web Push notifications to the nominated peer.
5. Upon the peer's confirmation in SIAKAD, the database updates duty rosters automatically.

### Running Dynamic QR Attendance
1. Before lab start, the on-duty assistant opens `https://siakad.algohub.web.id/buat-qr` on the projection computer.
2. Select the class section and click **Display QR**.
3. Project the rotating Dynamic QR Code for student mobile scanning.
4. Monitor successful check-ins under **Attendance Validation** (`/validasi-absensi`). Assistants can manually record attendance for students with camera hardware issues.
