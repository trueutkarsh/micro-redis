package microredis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func MarshalResp(i interface{}) string {
	switch i.(type) {
	case int64:
		return fmt.Sprintf(":%d\r\n", i)
	case int:
		return fmt.Sprintf(":%d\r\n", i)
	case string:
		return fmt.Sprintf("$%d\r\n%s\r\n", len(i.(string)), i.(string))
	case Key:
		return fmt.Sprintf("$%d\r\n%s\r\n", len(i.(string)), i.(string))
	case []string:
		result := fmt.Sprintf("*%d\r\n", len(i.([]string)))
		for _, k := range i.([]string) {
			result = result + fmt.Sprintf("$%d\r\n%s\r\n", len(k), k)
		}
		return result
	case []Key:
		result := fmt.Sprintf("*%d\r\n", len(i.([]Key)))
		for _, k := range i.([]Key) {
			result = result + fmt.Sprintf("$%d\r\n%s\r\n", len(k), k)
		}
		return result
	case error:
		result := fmt.Sprintf("-%s\r\n", i.(error).Error())
		return result
	default:
		if i == nil {
			return "$-1\r\n"
		} else {
			return fmt.Sprintf("-%s:%s\r\n", "Internal Server Error", fmt.Sprintf("Failed to perform marshalling to this type %v", i))
		}
	}
}

func UnmarshalResp(s string) ([]string, error) {
	switch s[0] {
	case byte(':'):
		i := strings.Index(s, "\r\n")
		_, err := strconv.ParseInt(s[1:i], 10, 0)
		if err != nil {
			return []string{}, errors.New(fmt.Sprintf("Unable parse resp string %s", s[1:i]))
		} else {
			return []string{s[1:i]}, nil
		}
	case byte('$'):
		i := strings.Index(s, "\r\n")
		_, err := strconv.ParseInt(s[1:i], 10, 32)
		if err != nil {
			return []string{}, errors.New(fmt.Sprintf("Unable parse resp string %s", s[1:i]))
		}
		result := strings.Split(s[i:], "\r\n")
		return result, nil
	case byte('*'):
		i := strings.Index(s, "\r\n")
		_, err := strconv.ParseInt(s[1:i], 10, 32)
		if err != nil {
			return []string{}, errors.New(fmt.Sprintf("Unable parse resp string %s", s[1:i]))
		}
		// split by $
		result := []string{}
		split_by_dollar := strings.Split(s[i+2:], "$")
		for _, i := range split_by_dollar {
			// split by \r\n
			j := strings.Split(i, "\r\n")
			result = append(result, j[1])
		}
		return result, nil
	case byte('-'):
		i := strings.Index(s, "\r\n")
		return []string{}, errors.New(s[1:i])
	default:
		return []string{}, errors.New(fmt.Sprintf("Unable parse resp string %s", s))
	}
}
