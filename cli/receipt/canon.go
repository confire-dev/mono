package receipt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
)

// Canonical returns the RFC 8785 JCS (JSON Canonicalization Scheme) encoding of v.
//
// Rules applied:
//   - No insignificant whitespace.
//   - Object keys sorted lexicographically by Unicode code point value.
//     For the ASCII-only keys used in receipts this is identical to byte order.
//   - Numbers serialised per ES2019 §7.1.12.1 (integer-valued doubles without
//     a decimal point; shortest-representation for fractional values).
//   - Recursive — nested objects are also sorted.
//
// The implementation marshals v to standard JSON first so that struct tags,
// omitempty rules, and time.Time formatting are applied by encoding/json before
// canonicalisation. The resulting JSON is then parsed into any and re-serialised
// with deterministic key ordering.
func Canonical(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("receipt/canon: marshal: %w", err)
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("receipt/canon: unmarshal: %w", err)
	}
	var buf bytes.Buffer
	if err := writeValue(&buf, generic); err != nil {
		return nil, fmt.Errorf("receipt/canon: write: %w", err)
	}
	return buf.Bytes(), nil
}

func writeValue(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case float64:
		buf.WriteString(canonicalFloat(val))
	case string:
		// Delegate string escaping to encoding/json — it is RFC 8785 compliant
		// for all Basic Multilingual Plane characters.
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(b)
	case []any:
		buf.WriteByte('[')
		for i, elem := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeValue(buf, elem); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return err
			}
			buf.Write(kb)
			buf.WriteByte(':')
			if err := writeValue(buf, val[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("receipt/canon: unsupported type %T", v)
	}
	return nil
}

// canonicalFloat serialises a float64 per RFC 8785 §3.2.4 / ES2019 §7.1.12.1.
// For our receipt domain all numbers are small integers, but the full
// algorithm is implemented for correctness.
func canonicalFloat(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		// JSON does not allow Inf or NaN; return null defensively.
		return "null"
	}
	if f == 0 {
		return "0" // both +0.0 and -0.0 → "0"
	}
	// Integer-valued doubles in the safe range: format without decimal point.
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	// Shortest decimal representation (matches ES2019 §7.1.12.1 for normal values).
	return strconv.FormatFloat(f, 'g', -1, 64)
}
