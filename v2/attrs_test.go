package golog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAttr_String(t *testing.T) {
	type Args struct {
		attr Attr
	}
	type Expects struct {
		want string
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name:    "string-value",
			args:    Args{attr: String("k", "v")},
			expects: Expects{want: "k=v"},
		},
		{
			name:    "int-value",
			args:    Args{attr: Int("n", 3)},
			expects: Expects{want: "n=3"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expects.want, tc.args.attr.String(), "String()")
		})
	}
}

func TestAttr_Equal(t *testing.T) {
	type Args struct {
		a Attr
		b Attr
	}
	type Expects struct {
		wantEqual bool
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name: "same-key-and-value",
			args: Args{
				a: String("k", "v"),
				b: String("k", "v"),
			},
			expects: Expects{wantEqual: true},
		},
		{
			name: "different-value",
			args: Args{
				a: String("k", "v"),
				b: String("k", "x"),
			},
			expects: Expects{wantEqual: false},
		},
		{
			name: "different-key",
			args: Args{
				a: String("k", "v"),
				b: String("other", "v"),
			},
			expects: Expects{wantEqual: false},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expects.wantEqual, tc.args.a.Equal(tc.args.b), "Equal()")
		})
	}
}

func TestAttr_Float64(t *testing.T) {
	a := Float64("pi", 3.14)
	require.Equal(t, KindFloat64, a.Value.Kind())
	require.Equal(t, 3.14, a.Value.Float64())
}

func TestAttr_Group(t *testing.T) {
	a := Group("g", String("a", "1"), Int("b", 2))
	require.Equal(t, "g", a.Key)
	require.Equal(t, KindGroup, a.Value.Kind())
	attrs := a.Value.Group()
	require.Len(t, attrs, 2)
	require.Equal(t, "a", attrs[0].Key)
	require.Equal(t, "b", attrs[1].Key)
}

func TestAttr_Any(t *testing.T) {
	a := Any("m", map[string]int{"x": 1})
	require.Equal(t, "m", a.Key)
	require.Equal(t, KindAny, a.Value.Kind())
}

func TestAttr_Time(t *testing.T) {
	ts := time.Unix(1, 2).UTC()
	tm := Time("t", ts)
	require.Equal(t, KindTime, tm.Value.Kind())
}

func TestAttr_Duration(t *testing.T) {
	d := Duration("d", 5*time.Second)
	require.Equal(t, KindDuration, d.Value.Kind())
	require.Equal(t, 5*time.Second, d.Value.Duration())
}
