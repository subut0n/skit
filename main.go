package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/subut0n/skit/internal/ansi"
	"github.com/subut0n/skit/internal/config"
	"github.com/subut0n/skit/internal/detector"
	"github.com/subut0n/skit/internal/history"
	"github.com/subut0n/skit/internal/i18n"
	"github.com/subut0n/skit/internal/parser"
	"github.com/subut0n/skit/internal/ui"
)

// Version is set via -ldflags "-X main.Version=x.y.z"
var Version = "dev"

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, ansi.Red+format+ansi.Reset+"\n", args...)
	os.Exit(1)
}

// getPalette returns the ANSI color codes for the given color scheme.
func getPalette(scheme config.ColorScheme) []string {
	switch scheme {
	case config.ColorSchemeDeuteranopia:
		return []string{
			"\033[34m", "\033[33m", "\033[36m",
			"\033[35m", "\033[94m", "\033[93m",
		}
	case config.ColorSchemeTritanopia:
		return []string{
			"\033[31m", "\033[35m", "\033[32m",
			"\033[91m", "\033[95m", "\033[92m",
		}
	case config.ColorSchemeHighContrast:
		return []string{
			"\033[91m", "\033[93m", "\033[92m",
			"\033[96m", "\033[94m", "\033[95m",
		}
	default: // rainbow
		return []string{
			"\033[31m", "\033[33m", "\033[32m",
			"\033[36m", "\033[34m", "\033[35m",
		}
	}
}

func main() {
	// Parse flags: extract --root and -w/--workspace before the switch
	useRoot := false
	useWorkspace := false
	var filteredArgs []string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--root":
			useRoot = true
		case "-w", "--workspace":
			useWorkspace = true
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) > 0 {
		arg := filteredArgs[0]
		switch arg {
		case "--version", "-v":
			loadConfigAndSetLang()
			fmt.Printf(i18n.Get().VersionFormat+"\n", Version)
			return
		case "--help", "-h":
			cfg := loadConfigAndSetLang()
			printHelp(getPalette(cfg.Config.ColorScheme))
			return
		case "--config":
			runConfigSetup()
			return
		case "--lang":
			runLangSetup()
			return
		case "--colors":
			runColorSetup()
			return
		case "--keys":
			runKeysSetup()
			return
		case "--history", "-hist":
			loadConfigAndSetLang()
			showHistory()
			return
		case "init", "config":
			runConfigSetup()
			return
		default:
			if !strings.HasPrefix(arg, "-") {
				loadConfigAndSetLang()
				runDirectScript(arg, useRoot)
				return
			}
			cfg := loadConfigAndSetLang()
			printHelp(getPalette(cfg.Config.ColorScheme))
			os.Exit(1)
		}
	}

	// No arguments: launch the interactive menu
	cfg := loadConfigAndSetLang()

	// First launch: run the initial setup wizard
	if !cfg.Exists() {
		result, err := config.RunSetup()
		if err != nil {
			fatal(i18n.Get().ErrConfig, err)
		}
		cfg.Config.KeyScheme = result.KeyScheme
		cfg.Config.Language = result.Language
		cfg.Config.ColorScheme = result.ColorScheme
		if result.KeyScheme == config.KeySchemeCustom {
			upKey, downKey := captureCustomKeys()
			cfg.Config.CustomUpKey = upKey
			cfg.Config.CustomDownKey = downKey
		}
		_ = cfg.Save()
		fmt.Println()
	}

	// Resolve package.json path
	pkgPath := resolvePackageJSON(useRoot, useWorkspace)
	if pkgPath == "" {
		fatal("%s", i18n.Get().ErrNoPackageJSON)
	}

	m := i18n.Get()

	scripts, err := parser.Parse(pkgPath)
	if err != nil {
		fatal(m.ErrReadPackageJSON, err)
	}

	if len(scripts) == 0 {
		fatal("%s", m.ErrNoScripts)
	}

	// Detect package manager (look from the package.json dir, then walk up)
	pkgDir := filepath.Dir(pkgPath)
	pm := detector.Detect(pkgDir)
	// If no lockfile in the package dir, try the root
	if pm.Manager == detector.NPM && pkgDir != "" {
		rootPkg := parser.FindRootPackageJSON(pkgDir)
		if rootPkg != "" {
			rootPM := detector.Detect(filepath.Dir(rootPkg))
			if rootPM.Manager != detector.NPM {
				pm = rootPM
			}
		}
	}

	// Display context line: relative path + package manager
	printContext(pkgPath, pm)

	opts := ui.Options{
		KeyScheme:     cfg.Config.KeyScheme,
		ColorPalette:  getPalette(cfg.Config.ColorScheme),
		CustomUpKey:   cfg.Config.CustomUpKey,
		CustomDownKey: cfg.Config.CustomDownKey,
	}
	result := ui.Run(scripts, opts)

	if !result.Confirmed || result.Script == nil {
		fmt.Printf("%s%s%s\n", ansi.Gray, m.Cancelled, ansi.Reset)
		return
	}

	executeScript(result.Script.Name, result.Script.Command, pm)
}

// resolvePackageJSON determines which package.json to use based on flags.
func resolvePackageJSON(useRoot, useWorkspace bool) string {
	m := i18n.Get()

	if useWorkspace {
		return resolveWorkspace()
	}

	if useRoot {
		dir, err := os.Getwd()
		if err != nil {
			return ""
		}
		root := parser.FindRootPackageJSON(dir)
		if root != "" {
			fmt.Printf("%s%s%s\n", ansi.Gray, m.UsingRoot, ansi.Reset)
		}
		return root
	}

	return findPackageJSON()
}

// resolveWorkspace detects workspaces and lets the user pick one.
func resolveWorkspace() string {
	m := i18n.Get()
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	rootPkg := parser.FindRootPackageJSON(dir)
	if rootPkg == "" {
		fatal("%s", m.ErrNoPackageJSON)
	}

	workspaces := parser.ParseWorkspaces(rootPkg)
	if len(workspaces) == 0 {
		// No workspaces, fall back to root
		return rootPkg
	}

	fmt.Printf("%s%s%s%s\n\n", ansi.Bold, ansi.Purple, fmt.Sprintf(m.WorkspaceDetected, len(workspaces)), ansi.Reset)

	for i, ws := range workspaces {
		fmt.Printf("  %s%2d.%s %s%-30s%s %s%s%s\n",
			ansi.Purple, i+1, ansi.Reset,
			ansi.Bold, ws.Name, ansi.Reset,
			ansi.Gray, ws.Path, ansi.Reset,
		)
	}

	fmt.Printf("\n%s%s%s", ansi.Gray, m.WorkspacePrompt, ansi.Reset)

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" {
			fmt.Printf("%s%s%s\n", ansi.Gray, m.Cancelled, ansi.Reset)
			os.Exit(0)
		}
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(workspaces) {
			ws := workspaces[idx-1]
			fmt.Println()
			return ws.PkgPath
		}
		fmt.Printf("%s"+m.WorkspaceInvalid+"%s", ansi.Red, len(workspaces), ansi.Reset)
	}
}

// printContext displays the detected package.json path and package manager.
func printContext(pkgPath string, pm detector.Info) {
	m := i18n.Get()

	displayPath := pkgPath
	cwd, err := os.Getwd()
	if err == nil {
		if rel, err := filepath.Rel(cwd, pkgPath); err == nil {
			displayPath = rel
		}
	}

	// If relative path goes up (../), show absolute path with ~ instead
	if strings.HasPrefix(displayPath, "..") {
		abs, err := filepath.Abs(pkgPath)
		if err == nil {
			if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(abs, home) {
				displayPath = "~" + abs[len(home):]
			} else {
				displayPath = abs
			}
		}
	}

	// Compact display: path ▸ runner
	fmt.Printf("%s%s%s\n\n", ansi.Gray, fmt.Sprintf(m.ContextLine, displayPath, pm.Name), ansi.Reset)
}

// executeScript runs a script via the detected package manager.
func executeScript(name, command string, pm detector.Info) {
	m := i18n.Get()

	fmt.Printf("%s%s%s%s\n\n", ansi.Bold, ansi.Green, fmt.Sprintf(m.Executing, pm.RunCmd, name), ansi.Reset)

	args := strings.Fields(pm.RunCmd)
	args = append(args, name)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	hist, err := history.New()
	if err == nil {
		_ = hist.Add(name, command, pm.Name)
	}

	if err := cmd.Run(); err != nil {
		fatal("\n"+m.ErrCommandFailed, err)
	}

	fmt.Printf("\n%s%s%s%s\n", ansi.Bold, ansi.Green, m.Success, ansi.Reset)
}

func printHelp(palette []string) {
	type helpEntry struct {
		cmd  string
		desc string
	}
	entries := []helpEntry{
		{"skit", "Interactive menu"},
		{"skit <script>", "Run a script directly"},
		{"skit -w, --workspace", "Pick a workspace package"},
		{"skit --root", "Use root package.json"},
		{"skit --help, -h", "Show this help"},
		{"skit --version, -v", "Show version"},
		{"skit --config", "Configure language, colors and key scheme"},
		{"skit --lang", "Change language"},
		{"skit --colors", "Change color scheme"},
		{"skit --keys", "Change key scheme"},
		{"skit --history, -hist", "Show execution history"},
	}

	fmt.Printf("\n  %s%sskit%s %s— interactive script runner for package.json%s\n\n", ansi.Bold, ansi.Purple, ansi.Reset, ansi.Gray, ansi.Reset)
	fmt.Printf("  %sUsage:%s skit %s[options]%s %s[script]%s\n\n", ansi.Bold, ansi.Reset, ansi.Gray, ansi.Reset, ansi.Gray, ansi.Reset)

	for i, e := range entries {
		c := palette[i%len(palette)]
		fmt.Printf("    %s%-26s%s  %s%s%s\n", c, e.cmd, ansi.Reset, ansi.Gray, e.desc, ansi.Reset)
	}
	fmt.Println()
}

func loadConfigAndSetLang() *config.Manager {
	cfg, err := config.New()
	if err != nil {
		cfg = &config.Manager{Config: config.Config{
			KeyScheme:   config.KeySchemeArrows,
			Language:    i18n.LangEN,
			ColorScheme: config.ColorSchemeRainbow,
		}}
	}
	i18n.Set(cfg.Config.Language)
	return cfg
}

func withConfig(fn func(cfg *config.Manager)) {
	cfg, err := config.New()
	if err != nil {
		fatal(i18n.Get().ErrGeneric, err)
	}
	i18n.Set(cfg.Config.Language)
	fn(cfg)
	if err := cfg.Save(); err != nil {
		fatal(i18n.Get().ErrSaveConfig, err)
	}
}

func runConfigSetup() {
	withConfig(func(cfg *config.Manager) {
		result, err := config.RunSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.KeyScheme = result.KeyScheme
		cfg.Config.Language = result.Language
		cfg.Config.ColorScheme = result.ColorScheme
		if result.KeyScheme == config.KeySchemeCustom {
			upKey, downKey := captureCustomKeys()
			cfg.Config.CustomUpKey = upKey
			cfg.Config.CustomDownKey = downKey
		} else {
			cfg.Config.CustomUpKey = 0
			cfg.Config.CustomDownKey = 0
		}
	})
}

func runLangSetup() {
	withConfig(func(cfg *config.Manager) {
		lang, err := config.RunLangSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.Language = lang
	})
}

func runColorSetup() {
	withConfig(func(cfg *config.Manager) {
		cs, err := config.RunColorSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.ColorScheme = cs
	})
}

func runKeysSetup() {
	withConfig(func(cfg *config.Manager) {
		ks, err := config.RunKeysSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.KeyScheme = ks
		if ks == config.KeySchemeCustom {
			upKey, downKey := captureCustomKeys()
			cfg.Config.CustomUpKey = upKey
			cfg.Config.CustomDownKey = downKey
		} else {
			cfg.Config.CustomUpKey = 0
			cfg.Config.CustomDownKey = 0
		}
	})
}

func captureCustomKeys() (upKey, downKey byte) {
	m := i18n.Get()
	fmt.Println()

	fmt.Printf("  %s%s%s", ansi.Purple, m.ConfigKeyUpPrompt, ansi.Reset)
	up, err := ui.CaptureKey()
	if err != nil {
		fmt.Printf("\n%s%s%s\n", ansi.Gray, "  (raw mode unavailable, defaulting to z/s)", ansi.Reset)
		return 'z', 's'
	}
	fmt.Printf("%s%s%s\n", ansi.Bold, ui.KeyDisplayName(up), ansi.Reset)

	for {
		fmt.Printf("  %s%s%s", ansi.Purple, m.ConfigKeyDownPrompt, ansi.Reset)
		down, err := ui.CaptureKey()
		if err != nil {
			return up, 's'
		}
		if down == up {
			fmt.Printf("%s(same as up key, try again)%s\n", ansi.Red, ansi.Reset)
			continue
		}
		fmt.Printf("%s%s%s\n", ansi.Bold, ui.KeyDisplayName(down), ansi.Reset)

		upName := ui.KeyDisplayName(up)
		downName := ui.KeyDisplayName(down)
		fmt.Printf("\n%s%s%s%s\n", ansi.Bold, ansi.Green, fmt.Sprintf(m.ConfigKeyConfirmCustom, upName, downName), ansi.Reset)
		return up, down
	}
}

func runDirectScript(script string, useRoot bool) {
	pkgPath := resolvePackageJSON(useRoot, false)
	if pkgPath == "" {
		fatal("%s", i18n.Get().ErrNoPackageJSON)
	}

	m := i18n.Get()

	scripts, err := parser.Parse(pkgPath)
	if err != nil {
		fatal(m.ErrReadPackageJSON, err)
	}

	var found *parser.Script
	for _, s := range scripts {
		if s.Name == script {
			found = &s
			break
		}
	}

	if found == nil {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", ansi.Red, fmt.Sprintf(m.ErrUnknownScript, script), ansi.Reset)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", ansi.Gray, m.AvailableScripts, ansi.Reset)
		for _, s := range scripts {
			desc := s.Description
			if desc == "" {
				desc = s.Command
			}
			fmt.Fprintf(os.Stderr, "  %s•%s %s  %s%s%s\n", ansi.Purple, ansi.Reset, s.Name, ansi.Gray, desc, ansi.Reset)
		}
		os.Exit(1)
	}

	pkgDir := filepath.Dir(pkgPath)
	pm := detector.Detect(pkgDir)
	if pm.Manager == detector.NPM {
		rootPkg := parser.FindRootPackageJSON(pkgDir)
		if rootPkg != "" {
			rootPM := detector.Detect(filepath.Dir(rootPkg))
			if rootPM.Manager != detector.NPM {
				pm = rootPM
			}
		}
	}
	printContext(pkgPath, pm)
	executeScript(found.Name, found.Command, pm)
}

func findPackageJSON() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return parser.FindPackageJSON(dir)
}

func showHistory() {
	m := i18n.Get()
	hist, err := history.New()
	if err != nil {
		fatal(m.ErrReadHistory, err)
	}

	entries := hist.Recent(20)
	if len(entries) == 0 {
		fmt.Printf("%s%s%s\n", ansi.Gray, m.HistoryEmpty, ansi.Reset)
		return
	}

	fmt.Printf("%s%s%s%s\n\n", ansi.Bold, ansi.Purple, m.HistoryTitle, ansi.Reset)
	for i, e := range entries {
		age := formatAge(e.Timestamp)
		fmt.Printf("  %s%2d.%s  %s%-20s%s  %s%s%s  %s%s  %s%s\n",
			ansi.Purple, i+1, ansi.Reset,
			ansi.Bold, e.Script, ansi.Reset,
			ansi.Cyan, e.Runner, ansi.Reset,
			ansi.Gray, age,
			e.Directory, ansi.Reset,
		)
	}
}

func formatAge(t time.Time) string {
	m := i18n.Get()
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return m.TimeJustNow
	case d < time.Hour:
		return fmt.Sprintf(m.TimeMinutesAgo, int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf(m.TimeHoursAgo, int(d.Hours()))
	default:
		return t.Format("02/01 15:04")
	}
}
