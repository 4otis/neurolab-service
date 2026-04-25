package handler

import (
	"encoding/json"
	"net/http"

	"github.com/4otis/neurolab-service/internal/cases"
	"github.com/4otis/neurolab-service/internal/dto/request"
	"github.com/4otis/neurolab-service/internal/dto/response"
	"go.uber.org/zap"
)

type TeacherHandler struct {
	logger            *zap.Logger
	uploadUseCase     cases.UploadUseCase
	teacherLabUseCase cases.TeacherLabUseCase
}

func NewTeacherHandler(
	logger *zap.Logger,
	uploadUseCase cases.UploadUseCase,
	teacherLabUseCase cases.TeacherLabUseCase,
) *TeacherHandler {
	return &TeacherHandler{
		logger:            logger,
		uploadUseCase:     uploadUseCase,
		teacherLabUseCase: teacherLabUseCase,
	}
}

// UploadScript загружает ZIP архив со скриптом проверки
// @Summary      Загрузить скрипт проверки
// @Description  Загрузить ZIP архив со скриптом для проверки лабораторных работ
// @Tags         scripts
// @Accept       multipart/form-data
// @Produce      json
// @Param        script_name     formData  string  false "Имя скрипта (по умолчанию script.zip)"
// @Param        script.zip      formData  file    true  "ZIP архив со скриптом проверки"
// @Success      201             {object}  response.UploadScriptResponse
// @Failure      400             {string}  string  "Неверный формат данных"
// @Failure      413             {string}  string  "Файл слишком большой (максимум 50MB)"
// @Failure      500             {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/v1/scripts/upload [post]
func (h *TeacherHandler) UploadScript(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "file too large or invalid form data", http.StatusBadRequest)
		return
	}

	scriptFile, _, err := r.FormFile("script.zip")
	if err != nil {
		http.Error(w, "script.zip is required", http.StatusBadRequest)
		return
	}
	defer scriptFile.Close()

	scriptName := r.FormValue("script_name")
	if scriptName == "" {
		scriptName = "script.zip"
	}

	err = h.uploadUseCase.UploadScript(r.Context(), scriptName, scriptFile)
	if err != nil {
		h.logger.Error("failed to upload script",
			zap.String("script_name", scriptName),
			zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("script uploaded successfully",
		zap.String("script_name", scriptName))

	resp := response.UploadScriptResponse{
		Status:     "ok",
		Message:    "Script uploaded successfully",
		ScriptName: scriptName,
		ScriptPath: "/scripts/" + scriptName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TeacherHandler) GenerateLab(w http.ResponseWriter, r *http.Request) {
	var req request.GenerateLabRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Subject == "" || req.Topic == "" || req.Title == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	res, err := h.teacherLabUseCase.GenerateLab(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to generate lab", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
