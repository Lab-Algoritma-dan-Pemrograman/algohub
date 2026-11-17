package dto

import "time"

// Bentuk GameQuestion di gameService.ts.
type GameQuestion struct {
	ID          string  `json:"id"`
	Language    string  `json:"language"`
	Difficulty  *string `json:"difficulty"`
	Title       *string `json:"title"`
	Code        *string `json:"code"`
	BugLine     *int    `json:"bugLine"`
	Explanation *string `json:"explanation"`
}

type GameQuestionInput struct {
	Language    *string `json:"language"`
	Difficulty  *string `json:"difficulty"`
	Title       *string `json:"title"`
	Code        *string `json:"code"`
	BugLine     *int    `json:"bugLine"`
	Explanation *string `json:"explanation"`
}

type GameQuestionsRequest struct {
	Questions []GameQuestionInput `json:"questions" binding:"required"`
}

type SeedResponse struct {
	Seeded bool `json:"seeded"`
}

// String kosong tidak ditulis, sama dengan `if (question.title)` di updateGameQuestion.
type GameQuestionUpdate struct {
	Language    string        `json:"language"`
	Difficulty  string        `json:"difficulty"`
	Title       string        `json:"title"`
	Code        string        `json:"code"`
	BugLine     Optional[int] `json:"bugLine"`
	Explanation string        `json:"explanation"`
}

// Nilai null dari DB dikirim apa adanya seperti getGameSettings.
type GameSettings struct {
	BugHuntActive        *bool `json:"bugHuntActive"`
	BugHuntCActive       *bool `json:"bugHuntCActive"`
	BugHuntPythonActive  *bool `json:"bugHuntPythonActive"`
	BugHuntWeeklyLimit   *int  `json:"bugHuntWeeklyLimit"`
	BugHuntQuestionCount int   `json:"bugHuntQuestionCount"`
}

type GameSettingsUpdate struct {
	BugHuntActive       Optional[bool] `json:"bugHuntActive"`
	BugHuntCActive      Optional[bool] `json:"bugHuntCActive"`
	BugHuntPythonActive Optional[bool] `json:"bugHuntPythonActive"`
	BugHuntWeeklyLimit  Optional[int]  `json:"bugHuntWeeklyLimit"`
}

type BugHuntAccess struct {
	Allowed   bool  `json:"allowed"`
	PlaysUsed int64 `json:"playsUsed"`
	Limit     int   `json:"limit"`
}

type GameHistoryRequest struct {
	GameType string `json:"gameType"`
	XPEarned int    `json:"xpEarned"`
}

// XP dan level adalah nilai optimistis saveGameHistory; untuk peran selain admin/kordas DB menimpanya dari sumber XP.
type GameHistoryResponse struct {
	XP         int       `json:"xp"`
	Level      int       `json:"level"`
	Streak     int       `json:"streak"`
	LastActive time.Time `json:"lastActive"`
	LeveledUp  bool      `json:"leveledUp"`
}

type IDResponse struct {
	ID string `json:"id"`
}
