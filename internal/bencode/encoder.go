package bencode

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// Marshal returns the bencoding of v.
func Marshal(v interface{}) ([]byte, error) {
    var buf bytes.Buffer
    if err := encode(&buf, v); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func encode(w *bytes.Buffer, v interface{}) error {
    switch val := v.(type) {
    case int:
        encodeInt(w, int64(val))
    case int64:
        encodeInt(w, val)
    case string:
        encodeString(w, val)
    case []byte:
        encodeString(w, string(val)) // treat []byte as string
    case []interface{}:
        if err := encodeList(w, val); err != nil {
            return err
        }
    case map[string]interface{}:
        if err := encodeDict(w, val); err != nil {
            return err
        }
    default:
        return fmt.Errorf("bencode: unsupported type %T", v)
    }
    return nil
}

func encodeInt(w *bytes.Buffer, val int64) {
    w.WriteByte('i')
    w.WriteString(strconv.FormatInt(val, 10))
    w.WriteByte('e')
}

func encodeString(w *bytes.Buffer, val string) {
    w.WriteString(strconv.Itoa(len(val)))
    w.WriteByte(':')
    w.WriteString(val)
}

func encodeList(w *bytes.Buffer, list []interface{}) error {
    w.WriteByte('l')
    for _, item := range list {
        if err := encode(w, item); err != nil {
            return err
        }
    }
    w.WriteByte('e')
    return nil
}

// encodeDict encodes a map ensuring keys are sorted lexicographically
func encodeDict(w *bytes.Buffer, dict map[string]interface{}) error {
    w.WriteByte('d')
    
    // Sort keys
    keys := make([]string, 0, len(dict))
    for k := range dict {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    for _, k := range keys {
        encodeString(w, k)
        if err := encode(w, dict[k]); err != nil {
            return err
        }
    }
    
    w.WriteByte('e')
    return nil
}
