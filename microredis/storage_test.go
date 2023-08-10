package microredis_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	m "github.com/trueutkarsh/micro-redis/microredis"
)

func TestGet(t *testing.T) {
	s := m.NewStorage(time.Second)

	// base case set key, key exists
	s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	act := s.Get(m.Key("hello"))
	assert.Equal(t, *act, "world")

	// key does not exists
	act = s.Get(m.Key("hello2"))
	assert.Nil(t, act)

}

func TestSet(t *testing.T) {
	s := m.NewStorage(time.Second)

	// set key val, check val and expiry
	success, old_val := s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	assert.True(t, success)
	assert.Nil(t, old_val)
	assert.Equal(t, s.TTL(m.Key("hello")), int64(-1)) // no expiry

	exp := time.Now().Add(time.Hour)
	// set with new val and expiry, get old val, only set if key exists

	// first fail because key exists and argument passed is set_if_not_exists
	success, old_val = s.Set(m.Key("hello"), "world", &exp, true, false, false, true)
	assert.False(t, success)
	// Note we return nil even though old val might exist due to failed set operation
	assert.Nil(t, old_val)

	// operation succeeds now
	success, old_val = s.Set(m.Key("hello"), "world2", &exp, true, false, true, false)
	assert.True(t, success)
	assert.Equal(t, "world", *old_val)

	// set another key only if it doesn not exists and check old_val is nil
	success, old_val = s.Set(m.Key("bella"), "ciao", nil, true, false, false, true)
	assert.True(t, success)
	assert.Nil(t, old_val)

}

func TestDel(t *testing.T) {
	s := m.NewStorage(time.Second)

	s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	s.Set(m.Key("greetings"), "earth", nil, false, false, false, false)

	result := s.Del([]m.Key{"hello", "greetings", "bella"})
	// two keys deleted, other key didn't exist
	assert.Equal(t, 2, result)

}

func TestExpireKeyNotPresent(t *testing.T) {
	s := m.NewStorage(time.Second)
	result := s.Expire("hello", 10, false, false, false, false)
	// key does not exists
	assert.Equal(t, 0, result)
}

func TestExpireExpiryNotExists(t *testing.T) {
	s := m.NewStorage(time.Second)
	// case 1: set expiry if expiry is not set
	s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	result := s.Expire(m.Key("hello"), 60, true, false, false, false)
	assert.Equal(t, 1, result)
}

func TestExpireExpiryExists(t *testing.T) {
	s := m.NewStorage(time.Second)
	// case 2: set expiry if expiry is set
	// case 2.1: first fail because it is not set
	s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	result := s.Expire(m.Key("hello"), 60, false, true, false, false)
	assert.Equal(t, 0, result)
	// case 2.2
	// Now succeed
	exp := time.Now().Add(time.Minute)
	s.Set(m.Key("hello"), "world", &exp, false, false, false, false)
	result = s.Expire(m.Key("hello"), 60, false, true, false, false)
	assert.Equal(t, 1, result)
}

func TestExpireExpiryGT(t *testing.T) {
	s := m.NewStorage(time.Second)
	// case 3: set expiry if new expiry greater then old
	// case 3.1 first fail because it is not set hence lives forever
	s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	result := s.Expire(m.Key("hello"), 60, false, false, true, false)
	assert.Equal(t, 0, result)
	// case3.2
	exp := time.Now().Add(time.Minute)
	s.Set(m.Key("hello"), "world", &exp, false, false, false, false)
	// ~120 seconds ahead
	result = s.Expire(m.Key("hello"), 120, false, false, true, false)
	assert.Equal(t, 1, result)
}

func TestExpireExpirtLt(t *testing.T) {
	s := m.NewStorage(time.Second)
	// case 4: set expiry if new expiry is less than old
	// case 4.1: First fail because old expiry is ahead
	exp := time.Now().Add(2 * time.Minute)
	s.Set(m.Key("hello"), "world", &exp, false, false, false, false)
	// ~60 seconds ahead so fail
	result := s.Expire(m.Key("hello"), 180, false, false, false, true)
	assert.Equal(t, 0, result)
	// case 4.3
	s.Set(m.Key("hello"), "world", &exp, false, false, false, false)
	// ~60 seconds before so succeed
	result = s.Expire(m.Key("hello"), 60, false, false, false, true)
	assert.Equal(t, 1, result)
}

func TestTTL(t *testing.T) {
	s := m.NewStorage(time.Second)
	exp := time.Now().Add(time.Minute)
	s.Set(m.Key("hello"), "world", &exp, false, false, false, false)
	ttl := s.TTL(m.Key("hello"))
	assert.True(t, ttl < 60)
	time.Sleep(5)
	assert.True(t, ttl > 50)
}

func TestKeys(t *testing.T) {
	s := m.NewStorage(time.Second)
	s.Set(m.Key("hello"), "world", nil, false, false, false, false)
	s.Set(m.Key("hell"), "world", nil, false, false, false, false)
	s.Set(m.Key("bella"), "ciao", nil, false, false, false, false)

	result, err := s.Keys("hell.*")
	if err != nil {
		t.Errorf("Err -> %v", err.Error())
	} else {
		assert.True(t, reflect.DeepEqual([]m.Key{"hello", "hell"}, result))
	}

	result, err = s.Keys(".ell.")
	if err != nil {
		t.Errorf("Err -> %v", err.Error())
	} else {
		assert.True(t, reflect.DeepEqual([]m.Key{"hello", "bella"}, result))
	}

}
