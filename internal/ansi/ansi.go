package ansi

const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Red        = "\033[38;2;255;0;102m"
	Green      = "\033[38;2;0;255;136m"
	Yellow     = "\033[38;2;255;238;0m"
	Blue       = "\033[38;2;0;221;255m"
	Purple     = "\033[38;2;221;68;255m"
	Cyan       = "\033[38;2;0;255;204m"
	Gray       = "\033[38;2;110;110;150m"
	ClearLine  = "\033[2K"
	Up         = "\033[1A"
	HideCursor = "\033[?25l"
	ShowCursor = "\033[?25h"
)
