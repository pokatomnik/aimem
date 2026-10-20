package colors

import "github.com/fatih/color"

func Primary() *color.Color {
	return color.New(color.FgCyan)
}

func Secondary() *color.Color {
	return color.New(color.FgMagenta)
}
