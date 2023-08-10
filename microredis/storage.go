package microredis

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Key type to denote key in the key value storage
type Key string

// Value type to denote the value in the key value storage
//
// For simplicity purpose we are only considering string datatype
type Value struct {
	val    *string
	expiry *time.Time // nil time denotes infinite expiry
}

// Storage will be the key value in-memory storage where
// data is the actual data stored and
// clear_freq is the freq at which expired keys will be cleared
type Storage struct {
	data       map[Key]Value
	clear_freq time.Duration
}

// NewStorage function to create initialize and return a pointer
// to the instance of Storage
func NewStorage(freq time.Duration) *Storage {
	result := Storage{
		data:       make(map[Key]Value),
		clear_freq: freq,
	}
	return &result
}

// Get function to get the value for a key if it exists
// It also clears the key from storage if it has expired
func (s *Storage) Get(key Key) *string {
	result, prs := s.data[key]
	if prs {
		if result.expiry != nil {
			if time.Now().After(*result.expiry) {
				delete(s.data, key)
				return nil
			}
		}
		return s.data[key].val
	}
	return nil
}

// Set function to set a value for a corresponding key with expiry
// Various options explained as follows
// ret_old_val means return the old value
// keepttl means keep the existing ttl (if exists)
// set_if_exists means ONLY set the value if it already exists
// set_if_not_exists means ONLY set the value if it does NOT exists
func (s *Storage) Set(
	key Key,
	val string,
	exp *time.Time,
	ret_old_val bool,
	keep_ttl bool,
	set_if_exists bool,
	set_if_not_exists bool,
) (bool, *string) {

	if set_if_exists {
		return s.setIfKeyExists(key, val, exp, ret_old_val, keep_ttl)
	} else if set_if_not_exists {
		return s.setIfKeyNotExists(key, val, exp, ret_old_val, keep_ttl)
	} else {
		prev_val, prs := s.data[key]
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

// setIfKeyExists sets the key with value and exp
// IF the key already exists in the db. It can return
// old value or keep the existing expiry if the corresponding
// ret_old_val and keep_ttl flags are set.
func (s *Storage) setIfKeyExists(
	key Key,
	val string,
	exp *time.Time,
	ret_old_val bool,
	keep_ttl bool,
) (bool, *string) {
	prev_val, prs := s.data[key]
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

}

// setIfKeyNotExists sets the key with value and exp
// IF the key DOES NOT exists in the db. It can return
// old value or keep the existing expiry if the corresponding
// ret_old_val and keep_ttl flags are set.
func (s *Storage) setIfKeyNotExists(
	key Key,
	val string,
	exp *time.Time,
	ret_old_val bool,
	keep_ttl bool,
) (bool, *string) {
	prev_val, prs := s.data[key]
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
}

// Del function to del a set of keys from storage
// The return value denotes number of keys deleted
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

// Expire function to expire an existing key after
// certain number of secs
// ONLY one other condition should be present through
// arguments
// set_if_expiry means only set new expiry if not already set previously
// set_if_no_expiry means only set new expiry if it already exists
// set_if_gt means only set new expiry if it is greater than existing expiry
// set_if_lt means only set new expiry if it is less than existing expiry
// return 1 if setting new expiry is successful else 0 if it fails due to
// any of the args above
func (s *Storage) Expire(
	key Key,
	secs int64,
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
		return s.expireIfNoExpiry(key, secs)
	}

	if set_if_expiry {
		return s.expireIfExpiryExists(key, secs)
	}

	if set_if_gt {
		return s.expireIfExpiryGreater(key, secs)
	}

	if set_if_lt {
		return s.expireIfExpiryLesser(key, secs)
	}

	// base case
	if secs > 0 {
		new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
		val.expiry = &new_exp
		s.data[key] = val
		return 1
	} else {
		delete(s.data, key)
		return 0
	}
}

// expireIfNoExpire function sets expiry of a key in secs
// ONLY if expiry is not already set.
// Note: negative secs means clear the key from db
func (s *Storage) expireIfNoExpiry(key Key, secs int64) int {
	val, prs := s.data[key]
	if !prs {
		return 0
	}
	if val.expiry == nil {
		if secs > 0 {
			new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
			val.expiry = &new_exp
			s.data[key] = val
		} else {
			// delete the key
			delete(s.data, key)
		}
		return 1
	} else {
		return 0
	}
}

// expireIfExpiryExists function sets expiry of a key in secs ONLY
// if expiry already exists
// Note: negative secs means clear the key from db
func (s *Storage) expireIfExpiryExists(key Key, secs int64) int {
	val, prs := s.data[key]
	if !prs {
		return 0
	}
	if val.expiry != nil {
		if secs > 0 {
			new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
			val.expiry = &new_exp
			s.data[key] = val
		} else {
			delete(s.data, key)
		}
		return 1
	} else {
		return 0
	}
}

// expireIfExpiryGreater function sets expiry of a key in secs ONLY
// if new expiry is further in time than existing expiry
// Note: negative secs means clear the key from db
func (s *Storage) expireIfExpiryGreater(key Key, secs int64) int {
	val, prs := s.data[key]
	if !prs {
		return 0
	}
	if val.expiry != nil {
		if secs > 0 {
			new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
			if val.expiry.Before(new_exp) {
				val.expiry = &new_exp
				s.data[key] = val
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

// expireIfExpiryLesser function sets expiry of a key in secs ONLY
// if new expiry is before in time than existing expiry
// Note: negative secs means clear the key from db
func (s *Storage) expireIfExpiryLesser(key Key, secs int64) int {
	val, prs := s.data[key]
	if !prs {
		return 0
	}
	if val.expiry != nil {
		if secs > 0 {
			new_exp := time.Now().Add(time.Duration(float64(secs) * float64(time.Second)))
			if val.expiry.After(new_exp) {
				val.expiry = &new_exp
				s.data[key] = val
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

// TTL function to return the time to live of key in seconds
// returns -1 if key has no expiry and -2 if it does not exists
// This function also clears the existing key if it has expired
func (s *Storage) TTL(key Key) int64 {
	val, prs := s.data[key]
	if prs {
		if val.expiry == nil {
			return -1
		} else {
			result := int64(val.expiry.Sub(time.Now()).Seconds())
			// negative ttl implies key has expired so clear it
			if result < 0 {
				delete(s.data, key)
				return -2
			} else {
				return result
			}
		}
	} else {
		return -2
	}
}

// Keys function filters keys in the storage through a
// regex pattern and returns the arrays of keys that match
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

// ClearExpiredKeys function clears the all the keys in storage
// which have expired. This function is suppose the be run periodically.
// Note: this function has to be executed in a mutually exclusive way
// from other functions otherwise might lead to race condition or unexpected
// behaviour
func (s *Storage) ClearExpiredKeys() {
	for k, v := range s.data {
		if v.expiry != nil {
			if v.expiry.Before(time.Now()) {
				delete(s.data, k)
			}
		}
	}
}
