package i18n

import "testing"

func TestDefaultLanguageIsEN(t *testing.T) {
	m := Get()
	if m.VersionFormat != "skit version %s" {
		t.Errorf("default language should be EN, got VersionFormat=%q", m.VersionFormat)
	}
}

func TestSetFR(t *testing.T) {
	Set(LangFR)
	defer Set(LangEN)

	m := Get()
	if m.Cancelled != "Annulé." {
		t.Errorf("FR: Cancelled = %q, want 'Annulé.'", m.Cancelled)
	}
}

func TestSetES(t *testing.T) {
	Set(LangES)
	defer Set(LangEN)

	m := Get()
	if m.Cancelled != "Cancelado." {
		t.Errorf("ES: Cancelled = %q, want 'Cancelado.'", m.Cancelled)
	}
}

func TestSetDE(t *testing.T) {
	Set(LangDE)
	defer Set(LangEN)

	m := Get()
	if m.Cancelled != "Abgebrochen." {
		t.Errorf("DE: Cancelled = %q, want 'Abgebrochen.'", m.Cancelled)
	}
}

func TestSetUnknownDefaultsToEN(t *testing.T) {
	Set("xx")
	defer Set(LangEN)

	m := Get()
	if m.Cancelled != "Cancelled." {
		t.Errorf("unknown lang should default to EN, got Cancelled=%q", m.Cancelled)
	}
}

func TestSupportedLangs(t *testing.T) {
	langs := SupportedLangs()
	if len(langs) != 4 {
		t.Fatalf("expected 4 languages, got %d", len(langs))
	}
}
