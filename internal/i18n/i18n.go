package i18n

import "sync"

// Lang represents a supported language.
type Lang string

const (
	LangEN Lang = "en"
	LangFR Lang = "fr"
	LangES Lang = "es"
	LangDE Lang = "de"
)

// Messages holds all translatable strings used across the application.
type Messages struct {
	// main.go
	ErrConfig          string
	ErrNoPackageJSON   string
	PackageJSONFound   string
	ErrReadPackageJSON string
	ErrNoScripts       string
	Cancelled          string
	Executing          string
	ErrCommandFailed   string
	Success            string
	ErrGeneric         string
	ErrSaveConfig      string
	ErrReadHistory     string
	HistoryEmpty       string
	HistoryTitle       string
	TimeJustNow        string
	TimeMinutesAgo     string
	TimeHoursAgo       string
	ErrUnknownScript   string
	AvailableScripts   string
	VersionFormat      string
	DetectedRunner     string
	ContextLine        string // "%s  ▸  %s" (path, runner)
	WorkspaceDetected  string
	WorkspacePrompt    string
	WorkspaceInvalid   string
	UsingRoot          string

	// ui/menu.go
	MenuTitle         string
	FilterLabel       string
	FilterActiveLabel string
	NoMatchingScripts string
	ScriptCount       string
	HelpArrows        string
	HelpWASD          string
	HelpCustomFmt     string
	FallbackTitle     string
	FallbackPrompt    string
	FallbackInvalid   string

	// config.go — key scheme
	ConfigTitle            string
	ConfigKeyPrompt        string
	ConfigKeyArrows        string
	ConfigKeyArrowsHint    string
	ConfigKeyWASD          string
	ConfigKeyWASDHint      string
	ConfigKeyCustom        string
	ConfigKeyCustomHint    string
	ConfigKeyChoice        string
	ConfigKeyConfirmArrows string
	ConfigKeyConfirmWASD   string
	ConfigKeyConfirmCustom string
	ConfigKeyInvalid       string
	ConfigKeyUpPrompt      string
	ConfigKeyDownPrompt    string
	ConfigLangConfirm      string

	// config.go — color scheme
	ConfigColorPrompt           string
	ConfigColorRainbow          string
	ConfigColorRainbowHint      string
	ConfigColorDeuteranopia     string
	ConfigColorDeuteranopiaHint string
	ConfigColorTritanopia       string
	ConfigColorTritanopiaHint   string
	ConfigColorHighContrast     string
	ConfigColorHighContrastHint string
	ConfigColorChoice           string
	ConfigColorConfirm          string
	ConfigColorInvalid          string
}

var (
	current *Messages
	mu      sync.RWMutex
)

func init() {
	current = &messagesEN
}

// Set changes the active language.
func Set(lang Lang) {
	mu.Lock()
	defer mu.Unlock()
	switch lang {
	case LangFR:
		current = &messagesFR
	case LangES:
		current = &messagesES
	case LangDE:
		current = &messagesDE
	default:
		current = &messagesEN
	}
}

// Get returns the active Messages.
func Get() *Messages {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// SupportedLangs returns all supported language codes.
func SupportedLangs() []Lang {
	return []Lang{LangEN, LangFR, LangES, LangDE}
}
