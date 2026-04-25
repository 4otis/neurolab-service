package request

// Запрос на загрузку скрипта
type UploadScriptRequest struct {
	ScriptName string `json:"script_name"`
}

type GenerateLabRequest struct {
	Subject            string `json:"subject"`
	Topic              string `json:"topic"`
	Title              string `json:"title"`
	TeacherDescription string `json:"teacher_description"`
}
