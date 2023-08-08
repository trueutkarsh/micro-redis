package microredis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// MarshalResp function takes any valid object and based on its type
// converts it into a string
// Note: Since we are only dealing with String datatype as values and
// limited return value types from Storage class functions this function
// is minimial in its implementation
func MarshalResp(i interface{}) string {
	switch i.(type) {
	case int64:
		return fmt.Sprintf(":%d#", i)
	case int:
		return fmt.Sprintf(":%d#", i)
	case string:
		return fmt.Sprintf("$%d#%s#", len(i.(string)), i.(string))
	case Key:
		return fmt.Sprintf("$%d#%s#", len(i.(string)), i.(string))
	case []string:
		result := fmt.Sprintf("*%d#", len(i.([]string)))
		for _, k := range i.([]string) {
			result = result + fmt.Sprintf("$%d#%s#", len(k), k)
		}
		return result
	case []Key:
		result := fmt.Sprintf("*%d#", len(i.([]Key)))
		for _, k := range i.([]Key) {
			result = result + fmt.Sprintf("$%d#%s#", len(k), k)
		}
		return result
	case error:
		result := fmt.Sprintf("-%s#", i.(error).Error())
		return result
	default:
		if i == nil {
			return "$-1#"
		} else {
			return fmt.Sprintf("-%s:%s#", "Internal Server Error", fmt.Sprintf("Failed to perform marshalling to this type %v", i))
		}
	}
}

// UnmarshalResp function is opposite to MarshalResp function but rather
// than returning an interface{} object it returns an array of strings or error
// The values in these strings could be strings, int or nil (as these are the values)
// returned by our storage function
func UnmarshalResp(s string) ([]string, error) {
	switch s[0] {
	case byte(':'):
		i := strings.Index(s, "#")
		_, err := strconv.ParseInt(s[1:i], 10, 0)
		if err != nil {
			return []string{}, errors.New(fmt.Sprintf("Unable parse resp int %s", s[1:i]))
		} else {
			return []string{s[1:i]}, nil
		}
	case byte('$'):
		i := strings.Index(s, "#")
		val, err := strconv.ParseInt(s[1:i], 10, 32)
		if err != nil {
			return []string{}, errors.New(fmt.Sprintf("Unable parse resp string %s", s[1:i]))
		}
		// handle nil case
		if val == -1 {
			return []string{"nil"}, nil
		}
		result := strings.Split(s[i:], "#")
		return result, nil
	case byte('*'):
		i := strings.Index(s, "#")
		_, err := strconv.ParseInt(s[1:i], 10, 32)
		if err != nil {
			return []string{}, errors.New(fmt.Sprintf("Unable parse resp bulk %s", s[1:i]))
		}
		// split by $
		result := []string{}
		split_by_dollar := strings.Split(s[i+2:], "$")
		for _, i := range split_by_dollar {
			// split by #
			j := strings.Split(i, "#")
			result = append(result, j[1])
		}
		return result, nil
	case byte('-'):
		i := strings.Index(s, "#")
		return []string{}, errors.New(s[1:i])
	default:
		return []string{}, errors.New(fmt.Sprintf("Unable parse resp msg |%s|", s))
	}
}
