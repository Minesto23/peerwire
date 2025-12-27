package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Unmarshal parses the bencoded data and stores the result in the value pointed to by v.
// Supported types for v are:
//   - *int (or *int64, etc.)
//   - *string
//   - *[]interface{} (for lists)
//   - *map[string]interface{} (for dictionaries)
//   - interface{} (decodes into the most appropriate type: int64, string, []interface{}, map[string]interface{})
func Unmarshal(data []byte, v interface{}) error {
	r := bufio.NewReader(bytes.NewReader(data))
	val, err := decode(r)
	if err != nil {
		return err
	}
	
	// If v is nil or not a pointer, we can't do anything helpful (like json.Unmarshal)
	// For this scratch implementation, we'll support basic reflection-like setting
	// or just return the interface{} if v is *interface{}
	
	// Simplified reflection logic for the educational scope:
	switch target := v.(type) {
	case *interface{}:
		*target = val
	case *map[string]interface{}:
		m, ok := val.(map[string]interface{})
		if !ok {
			return errors.New("bencode: result is not a dictionary")
		}
		*target = m
	case *[]interface{}:
		l, ok := val.([]interface{})
		if !ok {
			return errors.New("bencode: result is not a list")
		}
		*target = l
    case *string:
        s, ok := val.(string)
        if !ok {
             return errors.New("bencode: result is not a string")
        }
        *target = s
    case *int:
        i, ok := val.(int64)
        if !ok {
            return errors.New("bencode: result is not an integer")
        }
        *target = int(i)
	default:
		return fmt.Errorf("bencode: unsupported type %T", v)
	}
	return nil
}

func decode(r *bufio.Reader) (interface{}, error) {
	b, err := r.Peek(1)
	if err != nil {
		return nil, err
	}

	switch {
	case b[0] == 'i':
		return decodeInt(r)
	case b[0] >= '0' && b[0] <= '9':
		return decodeString(r)
	case b[0] == 'l':
		return decodeList(r)
	case b[0] == 'd':
		return decodeDict(r)
	default:
		return nil, fmt.Errorf("bencode: invalid start character '%c'", b[0])
	}
}

func decodeInt(r *bufio.Reader) (int64, error) {
	// Consumes 'i'
	if _, err := r.ReadByte(); err != nil {
		return 0, err
	}
    
    // Read untill 'e'
	bytes, err := r.ReadBytes('e')
	if err != nil {
		return 0, err
	}
	
    // Remove 'e'
	numStr := string(bytes[:len(bytes)-1])
    
    // Check for negative zero or leading zeros (unless it's just "0")
    if len(numStr) > 1 && numStr[0] == '0' {
         return 0, errors.New("bencode: invalid integer (leading zero)")
    }
    if numStr == "-0" {
        return 0, errors.New("bencode: invalid integer (-0)")
    }

	return strconv.ParseInt(numStr, 10, 64)
}

func decodeString(r *bufio.Reader) (string, error) {
	// Read length prefix
	lenBytes, err := r.ReadSlice(':')
	if err != nil {
		return "", err
	}
    
    // Parse length (exclude ':')
    lenStr := string(lenBytes[:len(lenBytes)-1])
    strLen, err := strconv.ParseInt(lenStr, 10, 64)
    if err != nil {
        return "", err
    }
    
    if strLen < 0 {
        return "", errors.New("bencode: negative string length")
    }

	// Read exact bytes
	buf := make([]byte, strLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	
	return string(buf), nil
}

func decodeList(r *bufio.Reader) ([]interface{}, error) {
    // Consume 'l'
    if _, err := r.ReadByte(); err != nil {
        return nil, err
    }
    
    list := make([]interface{}, 0)
    
    for {
        b, err := r.Peek(1)
        if err != nil {
            return nil, err
        }
        
        if b[0] == 'e' {
            r.ReadByte() // Consume 'e'
            return list, nil
        }
        
        elem, err := decode(r)
        if err != nil {
            return nil, err
        }
        list = append(list, elem)
    }
}

func decodeDict(r *bufio.Reader) (map[string]interface{}, error) {
    // Consume 'd'
    if _, err := r.ReadByte(); err != nil {
        return nil, err
    }
    
    dict := make(map[string]interface{})
    
    for {
        b, err := r.Peek(1)
        if err != nil {
            return nil, err
        }
        
        if b[0] == 'e' {
            r.ReadByte() // Consume 'e'
            return dict, nil
        }
        
        // Keys must be strings
        keyInterface, err := decode(r)
        if err != nil {
            return nil, err
        }
        
        key, ok := keyInterface.(string)
        if !ok {
            return nil, errors.New("bencode: dictionary key must be a string")
        }
        
        val, err := decode(r)
        if err != nil {
            return nil, err
        }
        
        dict[key] = val
    }
}
