package main

import "github.com/caarlos0/env/v11"

type Config struct {
	LLMHost             string  `env:"AIMEM_LLM_HOST,notEmpty"`
	LLMAPIKey           string  `env:"AIMEM_LLM_API_KEY"`
	SummarizationModel  string  `env:"AIMEM_SUMMARIZATION_MODEL,notEmpty"`
	TopK                int     `env:"AIMEM_TOP_K" envDefault:"5"`
	SimilarityThreshold float32 `env:"AIMEM_SIMILARITY_THRESHOLD" envDefault:"0.7"`
	Dimensions          uint    `env:"AIMEM_DIMENSIONS,notEmpty"`
	WebHTTPHost         string  `env:"AIMEM_WEB_HTTP_HOST" envDefault:"127.0.0.1:8888"`
	QdrantHost          string  `env:"AIMEM_QDRANT_HOST"`
	QdrantAPIKey        string  `env:"AIMEM_QDRANT_API_KEY"`
	UsersJSON           string  `env:"AIMEM_USERS_JSON"`
	LogFile             string  `env:"LOG_FILE"`
	LogLevel            string  `env:"LOG_LEVEL" envDefault:"info"`
}

func GetConfig() (*Config, error) {
	return getConfig(nil)
}

// getConfig parses the process environment, or environ when it is not nil.
//
// The map form exists for tests: a test that called GetConfig would read the
// runner's environment, which the test rules forbid (tests are hermetic), and
// would silently change meaning depending on the shell that started it. An empty
// but non-nil map means "nothing is set", so a test can pin what a missing
// variable does, while nil keeps the production path reading os.Environ.
func getConfig(environ map[string]string) (*Config, error) {
	config := &Config{}

	opts := env.Options{}
	if environ != nil {
		opts.Environment = environ
	}

	if err := env.ParseWithOptions(config, opts); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetDimensions() uint32 {
	return uint32(c.Dimensions)
}

func (c *Config) GetLLMHost() string {
	return c.LLMHost
}

func (c *Config) GetLLMAPIKey() string {
	return c.LLMAPIKey
}

func (c *Config) GetSummarizationModel() string {
	return c.SummarizationModel
}

func (c *Config) GetTopK() int {
	return c.TopK
}

func (c *Config) GetWebHTTPHost() string {
	return c.WebHTTPHost
}

func (c *Config) GetSimilarityThreshold() float32 {
	return max(0, min(1, c.SimilarityThreshold))
}

func (c *Config) RequireAuth() bool {
	return c.QdrantHost != ""
}

func (c *Config) GetQdrantHost() string {
	return c.QdrantHost
}

func (c *Config) GetQdrantAPIKey() string {
	return c.QdrantAPIKey
}

func (c *Config) GetUsersJSON() string {
	return c.UsersJSON
}

func (c *Config) GetLogFile() string {
	return c.LogFile
}

func (c *Config) GetLogLevel() string {
	return c.LogLevel
}
