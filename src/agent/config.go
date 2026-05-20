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
	LogLevel      string
}

func NewConfigFromEnv(filePath string) (*Config, error) {

	// Load environment variables from the specified file
	// This will not override existing environment variables
	err := godotenv.Load(filePath)

	if err != nil {
		return nil, util.DetailedError("Unable to load environment vars", err)
	}

	// create config struct
	config := &Config{}

	// populate config struct with environment variables
	config.LLMApiBaseUrl = os.Getenv("LLM_BASE_URL")
	config.LLMApiKey = os.Getenv("LLM_API_KEY")
	config.LLMModel = os.Getenv("LLM_MODEL")
	config.LogLevel = os.Getenv("LOG_LEVEL")

	// set default log level if not provided
	if config.LogLevel == "" {
		config.LogLevel = "INFO"
	}

	// validate required environment variables
	if config.LLMApiBaseUrl == "" {
		return nil, util.DetailedError("Env variable LLM_BASE_URL not found", nil)
	}

	if config.LLMApiKey == "" {
		return nil, util.DetailedError("Env variable LLM_API_KEY not found", nil)
	}

	if config.LLMModel == "" {
		return nil, util.DetailedError("Env variable LLM_MODEL not found", nil)
	}

	return config, nil
}
