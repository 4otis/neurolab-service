package response

import "time"

// Ответ на загрузку решения лабораторной работы
type UploadSolutionResponse struct {
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Ответ со списком доступных лабораторных
type LabsListResponse struct {
	Labs []LabInfo `json:"labs"`
}

// Ответ с описанием лабораторной
type LabInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TestScript  string `json:"test_script"` //PIPELINE!
}

// Ответ со статусом проверки
type SubmissionStatusResponse struct {
	SubmissionID string    `json:"submission_id"`
	Status       string    `json:"status"`      // pending, building, testing, success, failed
	UploadedAt   time.Time `json:"uploaded_at"` // Время загрузки
}
