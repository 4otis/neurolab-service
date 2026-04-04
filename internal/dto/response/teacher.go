package response

// Ответ на загрузку скрипта
type UploadScriptResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	ScriptName string `json:"script_name"`
	ScriptPath string `json:"script_path"`
}
