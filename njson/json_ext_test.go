package njson

import (
	"strings"
	"testing"
)

func TestNwNodeGetInt64ByPath_largeIntegerNoFloatRounding(t *testing.T) {
	// 9007199254740993 > 2^53; float64 would round it to 9007199254740992.
	const want int64 = 9007199254740993
	root, err := NewNwNodeByJsonStr(`{"x":9007199254740993}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := root.GetInt64ByPath("x"); got != want {
		t.Fatalf("GetInt64ByPath: got %d want %d", got, want)
	}
}

func TestNwNodeGetInt64ByPath_scientificInteger(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"x":1e6}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := root.GetInt64ByPath("x"); got != 1_000_000 {
		t.Fatalf("got %d want 1000000", got)
	}
}

func TestNwNodeGetInt64ByPath_negativeBig(t *testing.T) {
	const want int64 = -9007199254740993
	root, err := NewNwNodeByJsonStr(`{"x":-9007199254740993}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := root.GetInt64ByPath("x"); got != want {
		t.Fatalf("got %d want %d", got, want)
	}
}

func TestNwNodeGetInt64ByPath_notIntegral(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"x":1.5}`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-integral value")
		}
	}()
	root.GetInt64ByPath("x")
}

func TestNwNodeGetFloat64ByPath(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"pi":3.141592653589793}`)
	if err != nil {
		t.Fatal(err)
	}
	got := root.GetFloat64ByPath("pi")
	if got < 3.141592653589 || got > 3.141592653589794 {
		t.Fatalf("unexpected float: %v", got)
	}
}

func TestNwNodeGetStringByPath_nested(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"a":{"b":"hello"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := root.GetStringByPath("a", "b"); got != "hello" {
		t.Fatalf("got %q want hello", got)
	}
}

func TestNwNodeGetNumberByPath_preservesLiteral(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"n":123456789012345678901234567890}`)
	if err != nil {
		t.Fatal(err)
	}
	got := root.getNumberByPath("n")
	// The literal is preserved verbatim, not rounded through float64.
	if !strings.HasPrefix(string(got), "123456789012345678901234567890") {
		t.Fatalf("literal changed: %s", got)
	}
}

func TestNwNodeGetBoolByPath(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"ok":true,"no":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if !root.GetBoolByPath("ok") {
		t.Fatal("expect true")
	}
	if root.GetBoolByPath("no") {
		t.Fatal("expect false")
	}
}

func TestNwNodeGetByArrayIndex(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"arr":[10,20,30]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := root.GetInt64ByPath("arr", 1); got != 20 {
		t.Fatalf("got %d want 20", got)
	}
}

func TestNwNodeToString(t *testing.T) {
	root, err := NewNwNodeByJsonStr(`{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	s, err := root.toString()
	if err != nil {
		t.Fatal(err)
	}
	if s != `{"a":1}` {
		t.Fatalf("got %s", s)
	}
}
