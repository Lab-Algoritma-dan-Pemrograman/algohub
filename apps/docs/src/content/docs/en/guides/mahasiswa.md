---
title: Student Guide
description: Step-by-step walkthrough for students covering lab practicum participation, e-learning study routines, QR attendance, and report submission.
---

# Student Guide

This guide is designed for enrolled student practicum participants (*praktikan*) at the Algorithm & Programming Laboratory to navigate all laboratory coursework, self-paced learning, and administrative workflows in AlgoHub.

---

## 1. Account Access & Single Sign-On (SSO)

Students maintain a single unified account credential across the AlgoHub ecosystem linked directly to their **Student Identification Number (NIM)**.

### Accessing Systems:
1. **SIAKAD (`siakad.algohub.web.id`)**:
   - Log in using your NIM and assigned lab password.
   - For first-time logins, use the default credentials issued by lab coordinators, then immediately update your password under **Profile**.
2. **SSO into E-Learning**:
   - In the SIAKAD navigation drawer, click **E-Learning**.
   - The system initiates an automated single sign-on handshake (SSO) and delivers you directly to your personalized learning dashboard without prompt for re-authentication.
3. **Praktikum (`praktikum.algohub.web.id`)**:
   - Log in via `/praktikum/login` using your NIM and lab password.

---

## 2. Participating in Laboratory Practicum Sessions

### Checking the Lab Schedule
Prior to scheduled sessions, consult your class timetable at `https://praktikum.algohub.web.id/info/jadwal`. Ensure you arrive at the laboratory at least 10 minutes prior to session activation.

### Entering the Assessment Room (`/praktikum/ruang`)
1. Connect your workstation to the designated laboratory network.
2. Navigate to **Ruang Praktikum**. When the presiding teaching assistant opens the session, room status toggles to **Active**.
3. If token protection is active, input the 6-character room token projected on the lab screen.
4. Click **Enter Session**. The synchronized countdown timer begins running in the top-right corner.

### Solving Challenges & In-Browser Code Testing
Inside the room, you will encounter structured prompts (Pretests, Posttests, Skill Challenges, or Final Exams):
- **Monaco Code Editor**: Author your code directly in the integrated editor. Font sizing and editor themes can be toggled via preferences.
- **Run & Test**: Click **Run Code**. Your C or Python source code compiles and executes immediately inside your browser.
- **Interactive Terminal**: When programs prompt for standard input (`scanf` in C or `input()` in Python), enter responses directly into the dark terminal prompt and press `Enter`.
- **Autosave Verification**: Every keystroke is saved as a browser draft, then synced to the server at most every 15 seconds. Drafts survive a page refresh. Confirm that the status reads **"Tersimpan otomatis" (Saved automatically)**.

### Submitting Your Work
1. Review all answers before time expires.
2. Click **Submit Answers** at the bottom of the interface.
3. Confirm final submission. Once locked, submissions cannot be amended.
4. **Timer Expiration**: If the countdown hits `00:00:00`, the workspace locks and the last drafts are saved to the server. The backend finalizes the submission. If the final save fails, the screen shows a warning and the drafts remain on the device - do not close the browser or log out.
5. **If Submit is rejected**: While any answer has not reached the server, Submit is blocked with a warning. Wait for the connection to recover and click Submit again; your answers are not lost.

---

## 3. Comprehensive Final Report Submission

At semester conclusion, students submit their term software engineering documentation:
1. Open **Laporan Besar** at `/praktikum/laporan-besar`.
2. Click **Upload File**.
3. Select your project archive (accepted formats include `.pdf` or compressed `.zip` according to course specifications).
4. Click **Submit Report**.
5. Monitor submission review status:
   - **Submitted**: File received by storage bucket, awaiting assistant grading.
   - **Revision Required**: The reviewing assistant left comments. Upload revised files before the cutoff deadline.
   - **Accepted (Marked)**: Final report approved and grade recorded in the academic registry.

---

## 4. Gamified Learning in E-Learning

Deepen your programming fluency outside structured lab hours:

### Exploring Courses & Lessons
1. Navigate to `https://elearning.algohub.web.id/courses`.
2. Select your targeted **Level** and **Module** (structured progressively from foundational syntax to advanced memory structures).
3. Complete **Lessons**: Read rich-text guides, arrange visual flowchart puzzles, take multiple-choice concept checks, and solve coding exercises.

### Maintaining Your Daily Study Streak
- Complete at least one lesson or quiz daily.
- Your profile streak fire icon will increment daily. Avoid missing calendar days to prevent your streak from resetting to zero.

### Weekly Bug Hunt Challenges
- Visit the Mini-Games tab weekly.
- Locate and pinpoint the single erroneous line in a flawed C or Python program before the clock expires to harvest weekly bonus XP.

### Class & Global Leaderboards
- Track your academic standing on the **Leaderboard**.
- Compete constructively with peers for top cohort rank and XP milestones.

---

## 5. Attendance Verification in SIAKAD

When entering and exiting the physical computer laboratory:
1. Open `https://siakad.algohub.web.id/absensi` on your mobile phone or laptop.
2. Grant camera permissions when prompted.
3. Scan the projected **Dynamic QR Code** displayed on the laboratory screen.
4. Wait for audio/visual confirmation of attendance registration.
5. Review your attendance record under the history log to ensure your presence has been recorded in the database.
