package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/4otis/neurolab-service/internal/cases"
	"github.com/4otis/neurolab-service/internal/dto/response"
	"go.uber.org/zap"
)

type StudentHandler struct {
	logger *zap.Logger
	uc     cases.StudentUseCase
}

func NewStudentHandler(logger *zap.Logger, uc cases.StudentUseCase) *StudentHandler {
	return &StudentHandler{
		logger: logger,
		uc:     uc,
	}
}

// GetAvailableLabs возвращает список доступных лабораторных работ для студента
// @Summary      Получить доступные лабораторные
// @Description  Получить список лабораторных работ, доступных студенту по его курсам
// @Tags         student
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        student_id    path      string  true  "ID студента"
// @Success      200           {object}  response.LabsListResponse
// @Failure      400           {string}  string  "Неверный формат данных"
// @Failure      401           {string}  string  "Не авторизован"
// @Failure      404           {string}  string  "Студент не найден"
// @Failure      500           {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/student/{student_id}/labs [get]
func (h *StudentHandler) GetAvailableLabs(w http.ResponseWriter, r *http.Request) {
	studentID := r.PathValue("student_id")
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	labs, err := h.uc.GetAvailableLabs(r.Context(), studentID)
	if err != nil {
		h.logger.Error("failed to get available labs",
			zap.String("student_id", studentID),
			zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var labInfos []response.LabInfo
	for _, lab := range labs {
		labInfos = append(labInfos, response.LabInfo{
			ID:          lab.ID,
			Title:       lab.Title,
			Description: lab.Description,
			TestScript:  lab.TestScript,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.LabsListResponse{Labs: labInfos})
}

// UploadLab загружает ZIP архив с решением лабораторной работы
// @Summary      Загрузить решение лабораторной работы
// @Description  Загрузить ZIP архив с решением. Архив будет передан в Docker контейнер для тестирования
// @Tags         student
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        student_id      path      string  true  "ID студента"
// @Param        lab_id          formData  string  true  "ID лабораторной работы"
// @Param        solution.zip    formData  file    true  "ZIP архив с решением (макс. 100MB)"
// @Success      201             {object}  response.UploadResponse
// @Failure      400             {string}  string  "Неверный формат данных"
// @Failure      401             {string}  string  "Не авторизован"
// @Failure      404             {string}  string  "Студент или лабораторная не найдены"
// @Failure      413             {string}  string  "Файл слишком большой (максимум 100MB)"
// @Failure      500             {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/student/{student_id}/upload [post]
func (h *StudentHandler) UploadLab(w http.ResponseWriter, r *http.Request) {
	studentID := r.PathValue("student_id")
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "file too large or invalid form data", http.StatusBadRequest)
		return
	}

	labID := r.FormValue("lab_id")
	if labID == "" {
		http.Error(w, "lab_id is required", http.StatusBadRequest)
		return
	}

	zipFile, _, err := r.FormFile("solution.zip")
	if err != nil {
		http.Error(w, "solution.zip is required", http.StatusBadRequest)
		return
	}
	defer zipFile.Close()

	err = h.uc.UploadSolution(r.Context(), studentID, labID, zipFile)
	if err != nil {
		h.logger.Error("failed to upload lab",
			zap.String("student_id", studentID),
			zap.String("lab_id", labID),
			zap.Error(err))

		switch {
		case err.Error() == "student not found" || err.Error() == "lab not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	h.logger.Info("lab uploaded successfully",
		zap.String("student_id", studentID),
		zap.String("lab_id", labID))

	resp := response.UploadSolutionResponse{
		Status:     "ok",
		Message:    "Lab uploaded successfully",
		UploadedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetSubmissionResult возвращает результат проверки лабораторной работы
// @Summary      Получить результат проверки
// @Description  Получить статус по загруженной лабораторной работе
// @Tags         student
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        student_id      path      string  true  "ID студента"
// @Param        submission_id   query     string  true  "ID сдачи"
// @Success      200             {object}  response.SubmissionStatusResponse
// @Failure      400             {string}  string  "Неверный формат данных"
// @Failure      401             {string}  string  "Не авторизован"
// @Failure      403             {string}  string  "Доступ запрещен"
// @Failure      404             {string}  string  "Сдача не найдена"
// @Failure      500             {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/student/{student_id}/submission [get]
func (h *StudentHandler) GetSubmissionResult(w http.ResponseWriter, r *http.Request) {
	studentID := r.PathValue("student_id")
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	submissionID := r.URL.Query().Get("submission_id")
	if submissionID == "" {
		http.Error(w, "submission_id is required", http.StatusBadRequest)
		return
	}

	submission, err := h.uc.GetSubmission(r.Context(), submissionID)
	if err != nil {
		h.logger.Error("failed to get submission",
			zap.String("submission_id", submissionID),
			zap.Error(err))
		http.Error(w, "submission not found", http.StatusNotFound)
		return
	}

	if submission.StudentID != studentID {
		h.logger.Warn("access denied to submission",
			zap.String("student_id", studentID),
			zap.String("submission_owner", submission.StudentID),
			zap.String("submission_id", submissionID))
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	resp := response.SubmissionStatusResponse{
		SubmissionID: submission.ID,
		Status:       submission.Status,
		UploadedAt:   submission.SubmittedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
