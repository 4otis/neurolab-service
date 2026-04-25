package cases

import (
	"context"
	"fmt"
	"strings"

	"github.com/4otis/neurolab-service/internal/port/repo"
	"go.uber.org/zap"
)

const systemPrompt = `
Ты — опытный методист и преподаватель IT-дисциплин в ВУЗе.
Генерируешь полный курс лабораторных работ по заданной дисциплине.
Отвечай ТОЛЬКО валидным JSON. Никакого текста до или после JSON.
Не используй теги <think>. Не пиши рассуждений. Не добавляй пояснений.
Первый символ ответа — {, последний — }.
Думай кратко, отвечай сразу JSON.
`

func buildUserPrompt(p GenerateCourseParams) string {
	const tmpl = `
Сгенерируй курс лабораторных работ по дисциплине "{{subject}}".
Ответь строго в формате JSON:

{
  "course": {
    "subject": "Название дисциплины",
    "description": "Краткое описание курса — 2-3 предложения",
    "labs_count": число,
    "variants_count": число
  },

  "labs": [
    {
      "lab_number": 1,
      "title": "Название лабораторной работы",
      "topic": "Тема лабораторной",
      "goal": "Цель работы — одним предложением",
      "difficulty": "beginner" или "intermediate" или "advanced",
      "estimated_time_minutes": число,

      "common": {
        "theory_summary": "Краткая теоретическая справка, необходимая для выполнения",
        "tools": ["инструмент 1", "инструмент 2"],
        "evaluation_criteria": ["Критерий 1", "Критерий 2"]
      },

      "variants": [
        {
          "variant_number": 1,
          "title": "Подзаголовок варианта",
          "task_description": "Полная формулировка задания для студента",
          "input_data": "Входные данные или условия задачи, null если не применимо",
          "expected_output": "Что должно получиться на выходе",
          "hints": ["Подсказка 1", "Подсказка 2"],
          "extra_challenge": "Усложнённое задание для тех кто справился быстро"
        }
      ],

      "meta": {
        "reuse_warning": "ok" или "similar" или "duplicate"
      }
    }
  ]
}

Параметры:
  subject:        "{{subject}}"
  labs_count:     {{labs_count}}
  variants_count: {{variants_count}}
{{custom_notes_block}}

Правила генерации курса:
  — Сгенерируй ровно {{labs_count}} лабораторных работ
  — Лабораторные выстроены в логическую последовательность:
    каждая следующая опирается на знания из предыдущей
  — Сложность плавно нарастает от beginner к advanced
  — Темы лабораторных определи самостоятельно исходя из дисциплины,
    охвати ключевые разделы курса равномерно

Правила генерации вариантов:
  — В каждой лабораторной ровно {{variants_count}} вариантов
  — Варианты равнозначны по сложности, но различаются по содержанию:
    разные входные данные, предметные области или контекст задачи
  — Каждый вариант самодостаточен — студент не нуждается в других вариантах
  — Не дублировать формулировки дословно между вариантами
  — extra_challenge обязателен в каждом варианте

difficulty:
  "beginner":     задания на воспроизведение — студент повторяет по инструкции
  "intermediate": задания на применение — студент адаптирует знания к новой ситуации
  "advanced":     задания на анализ/синтез — студент самостоятельно проектирует решение

estimated_time_minutes:
  beginner:     30–60
  intermediate: 60–120
  advanced:     120–180

reuse_warning:
  "ok":        варианты достаточно различаются, списать сложно
  "similar":   варианты похожи, преподавателю стоит проверить вручную
  "duplicate": тема слишком узкая, варианты практически идентичны
`

	customNotesBlock := ""
	if strings.TrimSpace(p.CustomNotes) != "" {
		customNotesBlock = fmt.Sprintf("  custom_notes:   \"%s\"", p.CustomNotes)
	}

	return strings.NewReplacer(
		"{{subject}}", p.Subject,
		"{{labs_count}}", fmt.Sprintf("%d", p.LabsCount),
		"{{variants_count}}", fmt.Sprintf("%d", p.VariantsCount),
		"{{custom_notes_block}}", customNotesBlock,
	).Replace(tmpl)
}

type GenerateCourseParams struct {
	Subject       string
	LabsCount     int
	VariantsCount int
	CustomNotes   string
}

// type GenerateCourseParams struct {
// 	Subject        string
// 	Topic          string
// 	Difficulty     string // beginner | intermediate | advanced
// 	VariantsCount  int
// 	ExtraChallenge bool
// 	Language       string // ru | en
// 	CustomNotes    string
// }

var _ LLMUseCase = (*LLMUseCaseImpl)(nil)

type LLMUseCase interface {
	GenerateCourse(ctx context.Context, name, prompt string) (string, error)
	ApproveCourse(ctx context.Context, id string) error
	DenyCourse(ctx context.Context, id string) error
}

type LLMUseCaseImpl struct {
	logger  *zap.Logger
	llmRepo repo.LLMRepo
}

func NewLLMUseCase(
	logger *zap.Logger,
	llmRepo repo.LLMRepo,
) *LLMUseCaseImpl {
	return &LLMUseCaseImpl{
		logger:  logger,
		llmRepo: llmRepo,
	}
}

func (uc *LLMUseCaseImpl) GenerateCourse(ctx context.Context, name, prompt string) (answerJSON string, err error) {

	params := GenerateCourseParams{
		Subject:       "Мобильная разработка на Kotlin",
		LabsCount:     10,
		VariantsCount: 5,
		CustomNotes:   "",
	}
	userPrompt := buildUserPrompt(params)

	id, err := uc.llmRepo.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		uc.logger.Error("failed to Generate",
			zap.Error(err),
		)
		return "", fmt.Errorf("GenerateCourse: ask llm: %w", err)
	}

	uc.logger.Info("course succesfully generated",
		zap.String("id", id),
		zap.String("subject", params.Subject),
		zap.String("custom_notes", params.CustomNotes),
		zap.Int("vars", params.VariantsCount),
		zap.Int("labs", params.LabsCount),
	)

	return id, nil
}

func (uc *LLMUseCaseImpl) ApproveCourse(ctx context.Context, id string) error {
	return nil
}

func (uc *LLMUseCaseImpl) DenyCourse(ctx context.Context, id string) error {
	return nil
}
