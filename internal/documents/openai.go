package documents

import (
	"context"

	"github.com/tmc/langchaingo/llms/openai"
)

const EmbeddingModel = "text-embedding-3-small"
const Model = "gpt-4o-2024-05-13"

type OpenAI struct {
	APIKey string
	llm    *openai.LLM
}

func NewOpenAI(apiKey string) (*OpenAI, error) {
	llm, err := openai.New(
		openai.WithAPIType(openai.APITypeOpenAI),
		openai.WithModel(Model),
		openai.WithEmbeddingModel(EmbeddingModel),
	)
	if err != nil {
		return nil, err
	}
	return &OpenAI{
		llm:    llm,
		APIKey: apiKey,
	}, nil
}

func (o *OpenAI) CreateEmbedding(ctx context.Context, content string) ([]float32, error) {
	embeddings, err := o.llm.CreateEmbedding(ctx, []string{content})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}
