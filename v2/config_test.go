package golog

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLogger_unknownFormat(t *testing.T) {
	_, err := NewLogger(Config{Format: "bad", Output: ""})
	require.Error(t, err)
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
		{
			name: "text-stderr",
			args: Args{cfg: Config{Format: FormatText, Output: "stderr", Level: LevelDebug}},
			expects: Expects{wantErr: false},
		},
		{
			name: "source-field-without-enable-source",
			args: Args{cfg: Config{Format: FormatText, Output: "", Level: LevelDebug, SourceFieldName: "src"}},
			expects: Expects{wantErr: false},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLogger(tc.args.cfg)
			if tc.expects.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDevelopmentConfig(t *testing.T) {
	cfg := DevelopmentConfig()
	require.Equal(t, FormatText, cfg.Format)
	require.Equal(t, LevelDebug, cfg.Level)
	require.Equal(t, time.ANSIC, cfg.TimeFormat)
	require.True(t, cfg.EnableSource)
	require.Equal(t, "caller", cfg.SourceFieldName)
	require.Equal(t, SourceFormatFileLine, cfg.SourceFieldFormat)
	require.Equal(t, 2, cfg.SourceSkipFrames)
}

func TestProductionConfig(t *testing.T) {
	cfg := ProductionConfig()
	require.Equal(t, FormatJSON, cfg.Format)
	require.Equal(t, LevelInfo, cfg.Level)
	require.Equal(t, time.RFC3339Nano, cfg.TimeFormat)
	require.Equal(t, DurationFormatSeconds, cfg.DurationFormat)
}

func TestNewDevelopmentLogger(t *testing.T) {
	log, err := NewDevelopmentLogger()
	require.NoError(t, err)
	require.NotNil(t, log)
}

func TestNewProductionLogger(t *testing.T) {
	log, err := NewProductionLogger()
	require.NoError(t, err)
	require.NotNil(t, log)
}

func TestNewLogger_appendFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	log, err := NewLogger(Config{Format: FormatText, Output: path, Level: LevelDebug})
	require.NoError(t, err)
	require.NotNil(t, log)
	log.Info("line")
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestNewLogger_nilEnricherSkipped(t *testing.T) {
	log, err := NewLogger(Config{Format: FormatText, Output: "", Level: LevelDebug}, nil)
	require.NoError(t, err)
	require.NotNil(t, log)
}
