package agent

import (
	"agent0/util"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	LLMApiKey     string
	LLMApiBaseUrl string
	LLMModel      string
}

func NewConfigFromEnv(filePath string) (*Config, error) {
	err := godotenv.Load(filePath)
	if err != nil {
		return nil, util.NewErr("Unable to load environment vars", err)
	}
	config := &Config{}

	config.LLMApiBaseUrl = os.Getenv("LLM_BASE_URL")
	config.LLMApiKey = os.Getenv("LLM_API_KEY")
	config.LLMModel = os.Getenv("LLM_MODEL")

	if config.LLMApiBaseUrl == "" {
		return nil, util.NewErr("Env variable LLM_BASE_URL not found", nil)
	}

	if config.LLMApiKey == "" {
		return nil, util.NewErr("Env variable LLM_API_KEY not found", nil)
	}

	if config.LLMModel == "" {
		return nil, util.NewErr("Env variable LLM_MODEL not found", nil)
	}

	return config, nil
}
