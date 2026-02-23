package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/subut0n/skit/internal/ansi"
	"github.com/subut0n/skit/internal/config"
	"github.com/subut0n/skit/internal/i18n"
	"github.com/subut0n/skit/internal/parser"
)

// Options configures the interactive menu behavior.
type Options struct {
	KeyScheme     config.KeyScheme
	ColorPalette  []string
	CustomUpKey   byte
	CustomDownKey byte
}

// SelectionResult holds the user's script selection.
type SelectionResult struct {
	Script    *parser.Script
	Confirmed bool
}

// Run displays the interactive menu and returns the user's selection.
func Run(scripts []parser.Script, opts Options) SelectionResult {
	if len(scripts) == 0 {
		return SelectionResult{}
	}

	if opts.KeyScheme == "" {
		opts.KeyScheme = config.KeySchemeArrows
	}

	oldState, err := makeRaw()
	if err != nil {
		return runFallbackMenu(scripts, opts.ColorPalette)
	}
	defer restoreTerminal(oldState)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	defer close(done)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			restoreTerminal(oldState)
			fmt.Print(ansi.ShowCursor)
			os.Exit(130)
		case <-done:
		}
	}()

	fmt.Print(ansi.HideCursor)
	defer fmt.Print(ansi.ShowCursor)

	cursor := 0
	scroll := 0
	maxVisible := 15
	filter := ""
	filtering := false
	filtered := scripts
	prevLines := 0

	for {
		filtered = applyFilter(scripts, filter)
		if cursor >= len(filtered) {
			if len(filtered) == 0 {
				cursor = 0
			} else {
				cursor = len(filtered) - 1
			}
		}

		prevLines = renderMenu(filtered, cursor, scroll, maxVisible, filter, filtering, prevLines, opts)

		b := make([]byte, 4)
		n, err := os.Stdin.Read(b)
		if err != nil {
			clearLines(prevLines)
			return SelectionResult{}
		}
		if n == 0 {
			continue
		}
		key := b[:n]

		if filtering {
			switch {
			case key[0] == 27: // Escape
				filtering = false
				filter = ""
			case key[0] == 13: // Enter
				filtering = false
			case key[0] == 127 || key[0] == 8: // Backspace
				if len(filter) > 0 {
					filter = filter[:len(filter)-1]
				}
			case key[0] >= 32 && key[0] < 127:
				filter += string(key[0])
			}
			cursor = 0
			scroll = 0
			continue
		}

		switch {
		case isQuitKey(key[0], opts):
			clearLines(prevLines)
			return SelectionResult{}

		case key[0] == '/':
			filtering = true
			filter = ""

		case key[0] == 13: // Enter
			if len(filtered) == 0 {
				continue
			}
			selected := filtered[cursor]
			clearLines(prevLines)
			return SelectionResult{Script: &selected, Confirmed: true}

		case n >= 3 && key[0] == 27 && key[1] == 91:
			switch key[2] {
			case 65: // arrow up
				moveUp(&cursor, &scroll)
			case 66: // arrow down
				moveDown(&cursor, &scroll, maxVisible, len(filtered))
			}

		case isUpKey(key[0], opts):
			moveUp(&cursor, &scroll)

		case isDownKey(key[0], opts):
			moveDown(&cursor, &scroll, maxVisible, len(filtered))
		}
	}
}

func moveUp(cursor, scroll *int) {
	if *cursor > 0 {
		*cursor--
		if *cursor < *scroll {
			*scroll--
		}
	}
}

func moveDown(cursor, scroll *int, maxVisible, total int) {
	if *cursor < total-1 {
		*cursor++
		if *cursor >= *scroll+maxVisible {
			*scroll++
		}
	}
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func eqCaseInsensitive(a, b byte) bool {
	if isLetter(a) && isLetter(b) {
		return a|0x20 == b|0x20
	}
	return a == b
}

func isQuitKey(b byte, opts Options) bool {
	if b == 3 { // Ctrl+C
		return true
	}
	if opts.KeyScheme == config.KeySchemeCustom &&
		(eqCaseInsensitive(opts.CustomUpKey, 'q') || eqCaseInsensitive(opts.CustomDownKey, 'q')) {
		return false
	}
	return b == 'q' || b == 'Q'
}

func isUpKey(b byte, opts Options) bool {
	switch opts.KeyScheme {
	case config.KeySchemeWASD:
		return b == 'w' || b == 'W'
	case config.KeySchemeCustom:
		return eqCaseInsensitive(b, opts.CustomUpKey)
	}
	return false
}

func isDownKey(b byte, opts Options) bool {
	switch opts.KeyScheme {
	case config.KeySchemeWASD:
		return b == 's' || b == 'S'
	case config.KeySchemeCustom:
		return eqCaseInsensitive(b, opts.CustomDownKey)
	}
	return false
}

func helpLine(opts Options) string {
	m := i18n.Get()
	switch opts.KeyScheme {
	case config.KeySchemeWASD:
		return fmt.Sprintf("%s  %s%s", ansi.Gray, m.HelpWASD, ansi.Reset)
	case config.KeySchemeCustom:
		up := KeyDisplayName(opts.CustomUpKey)
		down := KeyDisplayName(opts.CustomDownKey)
		quitHint := "q quit"
		if opts.CustomUpKey == 'q' || opts.CustomDownKey == 'q' {
			quitHint = "Ctrl+C quit"
		}
		return fmt.Sprintf("%s  %s%s", ansi.Gray, fmt.Sprintf(m.HelpCustomFmt, strings.ToLower(up), strings.ToLower(down), quitHint), ansi.Reset)
	default:
		return fmt.Sprintf("%s  %s%s", ansi.Gray, m.HelpArrows, ansi.Reset)
	}
}

func renderMenu(scripts []parser.Script, cursor, scroll, maxVisible int, filter string, filtering bool, prevLines int, opts Options) int {
	for i := 0; i < prevLines; i++ {
		fmt.Print(ansi.Up + ansi.ClearLine)
	}

	lines := 0
	printLine := func(s string) {
		fmt.Println(s)
		lines++
	}

	msg := i18n.Get()
	printLine(fmt.Sprintf("%s%s%s%s", ansi.Bold, ansi.Purple, msg.MenuTitle, ansi.Reset))

	if filtering {
		printLine(fmt.Sprintf("%s  %s%s%s█%s", ansi.Gray, msg.FilterLabel, ansi.Reset, filter, ansi.Reset))
	} else if filter != "" {
		printLine(fmt.Sprintf("%s  %s%s%s%s", ansi.Gray, msg.FilterActiveLabel, ansi.Reset, filter, ansi.Reset))
	} else {
		printLine(helpLine(opts))
	}
	printLine("")

	if len(scripts) == 0 {
		printLine(fmt.Sprintf("%s  %s%s", ansi.Gray, msg.NoMatchingScripts, ansi.Reset))
	} else {
		end := scroll + maxVisible
		if end > len(scripts) {
			end = len(scripts)
		}

		// Calculate max name width for alignment
		maxNameLen := 0
		for i := scroll; i < end; i++ {
			if len(scripts[i].Name) > maxNameLen {
				maxNameLen = len(scripts[i].Name)
			}
		}
		if maxNameLen < 20 {
			maxNameLen = 20
		}

		fmtStr := fmt.Sprintf("%%-%ds", maxNameLen)

		for i := scroll; i < end; i++ {
			s := scripts[i]
			var line string
			if i == cursor {
				c := ansi.Purple
				if len(opts.ColorPalette) > 0 {
					c = opts.ColorPalette[i%len(opts.ColorPalette)]
				}
				line = fmt.Sprintf("  %s%s▶ %s%s"+fmtStr+"%s", ansi.Bold, ansi.Purple, ansi.Bold, c, s.Name, ansi.Reset)
			} else if len(opts.ColorPalette) > 0 {
				c := opts.ColorPalette[i%len(opts.ColorPalette)]
				line = fmt.Sprintf("    %s"+fmtStr+"%s", c, s.Name, ansi.Reset)
			} else {
				line = fmt.Sprintf("    "+fmtStr, s.Name)
			}

			// Show description (from x-skit) or command as description
			desc := s.Description
			if desc == "" {
				desc = s.Command
			}
			line += fmt.Sprintf("  %s%s%s", ansi.Gray, desc, ansi.Reset)

			printLine(line)
		}
		if len(scripts) > maxVisible {
			printLine(fmt.Sprintf("%s  "+msg.ScriptCount+"%s", ansi.Gray, cursor+1, len(scripts), ansi.Reset))
		}
	}

	return lines
}

func applyFilter(scripts []parser.Script, filter string) []parser.Script {
	if filter == "" {
		return scripts
	}
	f := strings.ToLower(filter)
	var result []parser.Script
	for _, s := range scripts {
		if strings.Contains(strings.ToLower(s.Name), f) ||
			strings.Contains(strings.ToLower(s.Command), f) ||
			strings.Contains(strings.ToLower(s.Description), f) {
			result = append(result, s)
		}
	}
	return result
}

func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print(ansi.Up + ansi.ClearLine)
	}
}

func runFallbackMenu(scripts []parser.Script, palette []string) SelectionResult {
	m := i18n.Get()
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s%s%s%s\n\n", ansi.Bold, ansi.Purple, m.FallbackTitle, ansi.Reset)
	for i, s := range scripts {
		numColor := ansi.Purple
		nameColor := ""
		nameReset := ""
		if len(palette) > 0 {
			c := palette[i%len(palette)]
			numColor = c
			nameColor = c
			nameReset = ansi.Reset
		}
		desc := s.Description
		if desc == "" {
			desc = s.Command
		}
		fmt.Printf("  %s%2d.%s %s%-30s%s %s%s%s\n", numColor, i+1, ansi.Reset, nameColor, s.Name, nameReset, ansi.Gray, desc, ansi.Reset)
	}
	fmt.Printf("\n%s%s%s", ansi.Gray, m.FallbackPrompt, ansi.Reset)

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" {
			return SelectionResult{}
		}
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(scripts) {
			s := scripts[idx-1]
			return SelectionResult{Script: &s, Confirmed: true}
		}
		fmt.Printf("%s"+m.FallbackInvalid+"%s", ansi.Red, len(scripts), ansi.Reset)
	}
}
