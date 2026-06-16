package receipt

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCanonical_SortsObjectKeys(t *testing.T) {
	// Map literal has non-deterministic Go iteration order, but Canonical must always
	// produce keys in lexicographic order regardless.
	input := map[string]any{
		"zebra":  1,
		"apple":  2,
		"mango":  3,
	}
	got, err := Canonical(input)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"apple":2,"mango":3,"zebra":1}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestCanonical_NestedObjectsSorted(t *testing.T) {
	input := map[string]any{
		"z": map[string]any{"b": 1, "a": 2},
		"a": map[string]any{"y": 3, "x": 4},
	}
	got, err := Canonical(input)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":{"x":4,"y":3},"z":{"a":2,"b":1}}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestCanonical_ArrayPreservesOrder(t *testing.T) {
	input := []any{"c", "a", "b"}
	got, err := Canonical(input)
	if err != nil {
		t.Fatal(err)
	}
	want := `["c","a","b"]`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestCanonical_NullValue(t *testing.T) {
	got, err := Canonical(nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "null" {
		t.Errorf("got %s, want null", got)
	}
}

func TestCanonical_Booleans(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, c := range cases {
		got, err := Canonical(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != c.want {
			t.Errorf("Canonical(%v) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestCanonical_IntegerFloat(t *testing.T) {
	// JSON numbers decode to float64 in Go; integer-valued floats must not
	// have a decimal point in canonical form.
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{-0.0, "0"},
		{1, "1"},
		{-1, "-1"},
		{42, "42"},
		{1000000, "1000000"},
	}
	for _, c := range cases {
		got := canonicalFloat(c.in)
		if got != c.want {
			t.Errorf("canonicalFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCanonical_StringEscaping(t *testing.T) {
	input := map[string]any{"key": "hello\nworld"}
	got, err := Canonical(input)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"key":"hello\nworld"}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestCanonical_EmptyObject(t *testing.T) {
	got, err := Canonical(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{}" {
		t.Errorf("got %s, want {}", got)
	}
}

func TestCanonical_TimeFieldIsString(t *testing.T) {
	// time.Time marshals to RFC 3339 string; canonical form must treat it as a string.
	ts := time.Date(2026, 6, 16, 10, 0, 1, 234000000, time.UTC)
	type wrapper struct {
		Timestamp time.Time `json:"timestamp"`
	}
	got, err := Canonical(wrapper{Timestamp: ts})
	if err != nil {
		t.Fatal(err)
	}
	// Verify it round-trips through JSON correctly.
	var check map[string]any
	if err := json.Unmarshal(got, &check); err != nil {
		t.Fatalf("canonical output is not valid JSON: %v", err)
	}
	if _, ok := check["timestamp"].(string); !ok {
		t.Error("expected timestamp to be a string in canonical JSON")
	}
}

func TestCanonical_DeterministicForSameStruct(t *testing.T) {
	// Two separately-constructed Body values with identical fields must produce
	// byte-for-byte identical canonical output.
	ts := time.Date(2026, 6, 16, 10, 0, 1, 0, time.UTC)
	b1 := Body{
		Version:             "1",
		ConfireVersion:      "1.4.2",
		ReceiptID:           "abc-123",
		PreviousPayloadHash: GenesisHash,
		Phase:               PhaseToolPre,
		Timestamp:           ts,
		Event: &Event{SessionID: "sess-1", Client: "claude_code"},
	}
	b2 := Body{
		Version:             "1",
		ConfireVersion:      "1.4.2",
		ReceiptID:           "abc-123",
		PreviousPayloadHash: GenesisHash,
		Phase:               PhaseToolPre,
		Timestamp:           ts,
		Event: &Event{SessionID: "sess-1", Client: "claude_code"},
	}
	c1, err := Canonical(b1)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Canonical(b2)
	if err != nil {
		t.Fatal(err)
	}
	if string(c1) != string(c2) {
		t.Errorf("same struct produced different canonical bytes:\n  c1=%s\n  c2=%s", c1, c2)
	}
}

func TestCanonical_OmitemptyFieldsAbsent(t *testing.T) {
	// Fields tagged omitempty must not appear in canonical JSON when zero-valued.
	b := Body{
		Version:             "1",
		ConfireVersion:      "1.0.0",
		ReceiptID:           "x",
		PreviousPayloadHash: GenesisHash,
		Phase:               PhaseSessionStart,
		Timestamp:           time.Now().UTC(),
		// ParentSessionID is empty → must be absent from canonical bytes.
		// Event, Decision, Provenance, Bundle are nil → must be absent.
	}
	got, err := Canonical(b)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, absent := range []string{"parent_session_id", "event", "decision", "provenance", "bundle"} {
		if _, ok := m[absent]; ok {
			t.Errorf("expected %q to be absent from canonical JSON, but it was present", absent)
		}
	}
}
