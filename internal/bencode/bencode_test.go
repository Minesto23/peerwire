package bencode

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer", "i42e", int64(42)},
		{"negative integer", "i-42e", int64(-42)},
		{"string", "4:spam", "spam"},
		{"list", "l4:spami42ee", []interface{}{"spam", int64(42)}},
		{"dictionary", "d3:bar4:spam3:fooi42ee", map[string]interface{}{"bar": "spam", "foo": int64(42)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal([]byte(tt.input), &result)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Unmarshal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected string
    }{
        {"integer", 42, "i42e"},
        {"string", "spam", "4:spam"},
        {"list", []interface{}{"spam", 42}, "l4:spami42ee"},
        {"dictionary", map[string]interface{}{"foo": 42, "bar": "spam"}, "d3:bar4:spam3:fooi42ee"}, // keys must be sorted
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Marshal(tt.input)
            if err != nil {
                t.Fatalf("Marshal() error = %v", err)
            }
            if string(got) != tt.expected {
                t.Errorf("Marshal() = %q, want %q", string(got), tt.expected)
            }
        })
    }
}
