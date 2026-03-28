package golog

import (
	"encoding/json"
	"testing"
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
			if got := tc.args.lvl.String(); got != tc.expects.want {
				t.Fatalf("got %q want %q", got, tc.expects.want)
			}
		})
	}
}

func TestLevel_String_unknownPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unknown level")
		}
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
			if err != nil {
				t.Fatalf("Marshal %v: %v", tc.args.lvl, err)
			}
			var got Level
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("Unmarshal %s: %v", b, err)
			}
			if got != tc.expects.want {
				t.Fatalf("round-trip: got %v want %v", got, tc.expects.want)
			}
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
			if err != nil {
				t.Fatalf("MarshalText %v: %v", tc.args.lvl, err)
			}
			var got Level
			if err := got.UnmarshalText(b); err != nil {
				t.Fatalf("UnmarshalText %q: %v", b, err)
			}
			if got != tc.expects.want {
				t.Fatalf("round-trip: got %v want %v", got, tc.expects.want)
			}
		})
	}
}

func TestLevel_UnmarshalText_caseInsensitive(t *testing.T) {
	var l Level
	if err := l.UnmarshalText([]byte("info")); err != nil {
		t.Fatal(err)
	}
	if l != LevelInfo {
		t.Fatalf("got %v", l)
	}
}
