package microredis_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/trueutkarsh/micro-redis/microredis"
)

func TestMarshalResp(t *testing.T) {
	var cases = []struct {
		input  interface{}
		result string
	}{
		{int64(123), ":123#"},
		{int(456), ":456#"},
		{"abc", "$3#abc#"},
		{microredis.Key("hello"), "$5#hello#"},
		{[]string{"key1", "key2"}, "*2#$4#key1#$4#key2#"},
		{[]microredis.Key{"key3", "key4"}, "*2#$4#key3#$4#key4#"},
		{errors.New("ERR: Test error"), "-ERR: Test error#"},
		{nil, "$-1#"},
	}

	for _, cs := range cases {

		testname := fmt.Sprintf("%v", cs.input)
		t.Run(testname, func(t *testing.T) {
			ans := microredis.MarshalResp(cs.input)
			if ans != cs.result {
				t.Errorf("got %v, want %v", ans, cs.result)
			}
		})
	}
}

func TestUnmarshalResp(t *testing.T) {
	var cases = []struct {
		input  string
		result []string
		err    error
	}{
		{":12345#", []string{"12345"}, nil},
		{"$5#hello#", []string{"hello"}, nil},
		{"$-1#", []string{"nil"}, nil},
		{"*2#$5#GET#$5#mykey#", []string{"GET", "mykey"}, nil},
		{"*3#$5#SET#$5#mykey#$5#value#", []string{"SET", "mykey", "value"}, nil},
		{"*2#$4#KEYS#$5#key.*#", []string{"KEYS", "key.*"}, nil},
		{"*2#$4#KEYS#$4#key*#", []string{"KEYS", "key*"}, nil},
		{"-Err: Custom Error#", []string{}, errors.New("Err: Custom Error")},
	}

	for _, cs := range cases {
		testname := fmt.Sprintf("%v", cs.input)
		t.Run(testname, func(t *testing.T) {
			ans, err := microredis.UnmarshalResp(cs.input)
			if err == nil {
				if !reflect.DeepEqual(ans, cs.result) {
					t.Errorf("got %v, want %v", ans, cs.result)
				}
			} else {
				if err.Error() != cs.err.Error() {
					t.Errorf("got %v, want %v", err, cs.err)
				}
			}
		})
	}
}
