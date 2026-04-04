package golog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLevel_String(t *testing.T) {
	type Args struct {
		lvl Level
	}
	type Expects struct {
		want string
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "debug", args: Args{lvl: LevelDebug}, expects: Expects{want: "DEBUG"}},
		{name: "info", args: Args{lvl: LevelInfo}, expects: Expects{want: "INFO"}},
		{name: "error", args: Args{lvl: LevelError}, expects: Expects{want: "ERROR"}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expects.want, tc.args.lvl.String())
		})
	}
}

func TestLevel_String_unknownPanics(t *testing.T) {
	defer func() {
		require.NotNil(t, recover(), "expected panic for unknown level")
	}()
	_ = Level(42).String()
}

func TestLevel_MarshalJSON_roundTrip(t *testing.T) {
	type Args struct {
		lvl Level
	}
	type Expects struct {
		want Level
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "debug", args: Args{lvl: LevelDebug}, expects: Expects{want: LevelDebug}},
		{name: "info", args: Args{lvl: LevelInfo}, expects: Expects{want: LevelInfo}},
		{name: "error", args: Args{lvl: LevelError}, expects: Expects{want: LevelError}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.args.lvl)
			require.NoError(t, err)
			var got Level
			err = json.Unmarshal(b, &got)
			require.NoError(t, err)
			require.Equal(t, tc.expects.want, got)
		})
	}
}

func TestLevel_MarshalText_roundTrip(t *testing.T) {
	type Args struct {
		lvl Level
	}
	type Expects struct {
		want Level
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "debug", args: Args{lvl: LevelDebug}, expects: Expects{want: LevelDebug}},
		{name: "info", args: Args{lvl: LevelInfo}, expects: Expects{want: LevelInfo}},
		{name: "error", args: Args{lvl: LevelError}, expects: Expects{want: LevelError}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.args.lvl.MarshalText()
			require.NoError(t, err)
			var got Level
			err = got.UnmarshalText(b)
			require.NoError(t, err)
			require.Equal(t, tc.expects.want, got)
		})
	}
}

func TestLevel_UnmarshalText_caseInsensitive(t *testing.T) {
	var l Level
	err := l.UnmarshalText([]byte("info"))
	require.NoError(t, err)
	require.Equal(t, LevelInfo, l)
}
