package microredis

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Define type key -> String
type Key string

// Define type value -> Struct (String, time)
type Value struct {
	val    *string
	expiry *time.Time // nil time denotes infinite expiry
}

type Storage struct {
	data       map[Key]Value
	clear_freq time.Duration
}

func NewStorage(freq time.Duration) *Storage {
	result := Storage{
		data:       make(map[Key]Value),
		clear_freq: freq,
	}
	return &result
}

func (s *Storage) Get(key Key) *string {
	result, prs := s.data[key]
	if prs {
		if time.Now().After(*result.expiry) {
			delete(s.data, key)
			return nil
		} else {
			return s.data[key].val
		}
	}
	return nil
}

func (s *Storage) Set(
	key Key,
	val string,
	exp *time.Time,
	ret_old_val bool,
	keep_ttl bool,
	set_if_exists bool,
	set_if_not_exists bool,
) (bool, *string) {

	prev_val, prs := s.data[key]
	if set_if_exists {
		if prs {
			// retain ttl
			if keep_ttl {
				exp = prev_val.expiry
			}
			s.data[key] = Value{
				val:    &val,
				expiry: exp,
			}
			if ret_old_val {
				return true, prev_val.val
			} else {
				return true, nil
			}
		} else {
			return false, nil
		}
	} else if set_if_not_exists {
		if !prs {
			// key not present retaining ttl makes no sense
			s.data[key] = Value{
				val:    &val,
				expiry: exp,
			}
			if ret_old_val {
				return true, prev_val.val
			} else {
				return true, nil
			}
		} else {
			return false, nil
		}
	} else {
		// if prs and keepttl retain ttl
		if prs && keep_ttl {
			exp = prev_val.expiry
		}
		s.data[key] = Value{
			val:    &val,
			expiry: exp,
		}
		if ret_old_val {
			return true, prev_val.val
		} else {
			return true, nil
		}
	}
}

func (s *Storage) Del(keys []Key) int {
	removed_count := 0
	for _, key := range keys {
		_, prs := s.data[key]
		if prs {
			delete(s.data, key)
			removed_count += 1
		}
	}
	return removed_count
}

func (s *Storage) Expire(
	key Key,
	secs int,
	set_if_no_expiry bool,
	set_if_expiry bool,
	set_if_gt bool,
	set_if_lt bool) int {
	// check arguments validity
	count := 0
	if set_if_expiry {
		count += 1
	}
	if set_if_no_expiry {
		count += 1
	}
	if set_if_gt {
		count += 1
	}
	if set_if_lt {
		count += 1
	}
	if count > 1 {
		// invalid arguments
		return 0
	}

	val, prs := s.data[key]

	// key does not exists
	if !prs {
		return 0
	}

	if set_if_no_expiry {
		if val.expiry == nil {

			if secs > 0 {
				new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
				val.expiry = &new_exp
			} else {
				// delete the key
				delete(s.data, key)
			}
			return 1
		} else {
			return 0
		}
	}

	if set_if_expiry {
		if val.expiry != nil {
			if secs > 0 {
				new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
				val.expiry = &new_exp
			} else {
				delete(s.data, key)
			}
			return 1
		} else {
			return 0
		}
	}

	if set_if_gt {
		if val.expiry != nil {
			if secs > 0 {
				new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
				if val.expiry.Before(new_exp) {
					val.expiry = &new_exp
					return 1
				} else {
					return 0
				}
			} else {
				delete(s.data, key)
				return 0
			}
		} else {
			return 0
		}
	}

	if set_if_lt {
		if val.expiry != nil {
			if secs > 0 {
				new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
				if val.expiry.After(new_exp) {
					val.expiry = &new_exp
					return 1
				} else {
					return 0
				}
			} else {
				delete(s.data, key)
				return 0
			}
		} else {
			return 0
		}
	}

	// base case
	if secs > 0 {
		new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
		val.expiry = &new_exp
		return 1
	} else {
		delete(s.data, key)
		return 0
	}
}

func (s *Storage) TTL(key Key) int {
	val, prs := s.data[key]
	if prs {
		if val.expiry == nil {
			return -1
		} else {
			return int(val.expiry.Sub(time.Now()).Seconds())
		}
	} else {
		return -2
	}
}

func (s *Storage) Keys(pattern string) ([]Key, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid pattern %s", pattern))
	}

	result := make([]Key, 0)

	for k := range s.data {
		if re.Match([]byte(k)) {
			result = append(result, k)
		}
	}
	return result, nil

}

// Basic expiry algorithm
func (s *Storage) ClearExpiredKeys() {
	for k, v := range s.data {
		if v.expiry.Before(time.Now()) {
			delete(s.data, k)
		}
	}
}
