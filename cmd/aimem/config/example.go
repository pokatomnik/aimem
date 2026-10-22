package config

import (
	_ "embed"
	"strings"

	"github.com/pokatomnik/gomem/internal/shared/colors"
)

//go:embed config.toml.example
var configExample string

func Example() string {
	var buf strings.Builder
	path, err := configPath()

	buf.WriteString(colors.Primary().Sprintln("CLI must be initialized first"))
	if err != nil {
		buf.WriteString(colors.Secondary().Sprintf("Create a config file: %s\n", path))
	}
	buf.WriteString(colors.Secondary().Sprintln("With content:"))
	buf.WriteString(colors.Secondary().Sprint(configExample))

	return buf.String()
}
