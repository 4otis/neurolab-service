package cases

import (
	"context"
	"strings"

	"github.com/4otis/neurolab-service/internal/clients"
	"github.com/4otis/neurolab-service/internal/dto/request"
	"github.com/4otis/neurolab-service/internal/dto/response"
)

var _ TeacherLabUseCase = (*TeacherLabUseCaseImpl)(nil)

type TeacherLabUseCase interface {
	GenerateLab(ctx context.Context, req request.GenerateLabRequest) (response.GenerateLabResponse, error)
}

type TeacherLabUseCaseImpl struct {
	llm *clients.Client
}

func NewTeacherLabUseCase(llm *clients.Client) *TeacherLabUseCaseImpl {
	return &TeacherLabUseCaseImpl{llm: llm}
}

func (uc *TeacherLabUseCaseImpl) GenerateLab(ctx context.Context, req request.GenerateLabRequest) (response.GenerateLabResponse, error) {
	prompt := buildPrompt(req)

	markdown, err := uc.llm.Generate(ctx, prompt)
	if err != nil {
		return response.GenerateLabResponse{}, err
	}

	return response.GenerateLabResponse{
		Markdown: markdown,
	}, nil
}

func buildPrompt(req request.GenerateLabRequest) string {
	var b strings.Builder

	b.WriteString("Ты — преподаватель по предмету «")
	b.WriteString(req.Subject)
	b.WriteString("».\n")
	b.WriteString("Сгенерируй лабораторную работу по теме «")
	b.WriteString(req.Topic)
	b.WriteString("».\n\n")

	if strings.TrimSpace(req.TeacherDescription) != "" {
		b.WriteString("Пожелания преподавателя:\n")
		b.WriteString(req.TeacherDescription)
		b.WriteString("\n\n")
	}

	b.WriteString("Требования:\n")
	b.WriteString("- Ответ строго на русском языке.\n")
	b.WriteString("- Формат: Markdown.\n")
	b.WriteString("- Не добавляй лишний текст вне структуры.\n\n")

	b.WriteString("# Лабораторная работа: ")
	b.WriteString(req.Title)
	b.WriteString("\n\n")
	b.WriteString("## Введение\n...\n")
	b.WriteString("## Цель работы\n...\n")
	b.WriteString("## Краткая теория\n...\n")
	b.WriteString("## Порядок выполнения\n...\n")
	b.WriteString("## Ожидаемые результаты\n...\n")
	b.WriteString("## Контрольные вопросы\n...\n")

	return b.String()
}
