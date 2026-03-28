package golog

import (
	"testing"
	"time"
)

func TestNewLogger_unknownFormat(t *testing.T) {
	_, err := NewLogger(Config{Format: "bad", Output: ""})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestNewLogger_validConfigs(t *testing.T) {
	type Args struct {
		cfg Config
	}
	type Expects struct {
		wantErr bool
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name: "text-stdout",
			args: Args{cfg: Config{Format: FormatText, Output: "", Level: LevelDebug}},
			expects: Expects{wantErr: false},
		},
		{
			name: "json-stdout",
			args: Args{cfg: Config{Format: FormatJSON, Output: "", Level: LevelDebug}},
			expects: Expects{wantErr: false},
		},
		{
			name: "text-enable-source",
			args: Args{cfg: Config{Format: FormatText, Output: "", Level: LevelDebug, EnableSource: true}},
			expects: Expects{wantErr: false},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLogger(tc.args.cfg)
			if (err != nil) != tc.expects.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.expects.wantErr)
			}
		})
	}
}

func TestDevelopmentConfig(t *testing.T) {
	cfg := DevelopmentConfig()
	if cfg.Format != FormatText || cfg.Level != LevelDebug || cfg.TimeFormat != time.ANSIC {
		t.Fatalf("DevelopmentConfig: %#v", cfg)
	}
	if !cfg.EnableSource || cfg.SourceFieldName != "caller" || cfg.SourceFieldFormat != SourceFormatFileLine || cfg.SourceSkipFrames != 2 {
		t.Fatalf("DevelopmentConfig source: %#v", cfg)
	}
}

func TestProductionConfig(t *testing.T) {
	cfg := ProductionConfig()
	if cfg.Format != FormatJSON || cfg.Level != LevelInfo || cfg.TimeFormat != time.RFC3339Nano {
		t.Fatalf("ProductionConfig: %#v", cfg)
	}
	if cfg.DurationFormat != DurationFormatSeconds {
		t.Fatalf("ProductionConfig DurationFormat: got %q", cfg.DurationFormat)
	}
}

func TestNewDevelopmentLogger(t *testing.T) {
	log, err := NewDevelopmentLogger()
	if err != nil {
		t.Fatal(err)
	}
	if log == nil {
		t.Fatal("nil logger")
	}
}

func TestNewProductionLogger(t *testing.T) {
	log, err := NewProductionLogger()
	if err != nil {
		t.Fatal(err)
	}
	if log == nil {
		t.Fatal("nil logger")
	}
}
