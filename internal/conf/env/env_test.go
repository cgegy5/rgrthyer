package env

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type subStruct struct {
	MyParam int
}

type mapEntry struct {
	MyValue  string
	MyStruct subStruct
}

type myDuration time.Duration

func (d *myDuration) UnmarshalJSON(b []byte) error {
	var in string
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}

	du, err := time.ParseDuration(in)
	if err != nil {
		return err
	}
	*d = myDuration(du)

	return nil
}

// UnmarshalEnv implements envUnmarshaler.
func (d *myDuration) UnmarshalEnv(s string) error {
	return d.UnmarshalJSON([]byte(`"` + s + `"`))
}

type testStruct struct {
	MyString     string
	MyInt        int
	MyFloat      float64
	MyBool       bool
	MyDuration   myDuration
	MyMap        map[string]*mapEntry
	MySlice      []string
	MySliceEmpty []string
}

func TestLoad(t *testing.T) {
	os.Setenv("MYPREFIX_MYSTRING", "testcontent")
	defer os.Unsetenv("MYPREFIX_MYSTRING")

	os.Setenv("MYPREFIX_MYINT", "123")
	defer os.Unsetenv("MYPREFIX_MYINT")

	os.Setenv("MYPREFIX_MYFLOAT", "15.2")
	defer os.Unsetenv("MYPREFIX_MYFLOAT")

	os.Setenv("MYPREFIX_MYBOOL", "yes")
	defer os.Unsetenv("MYPREFIX_MYBOOL")

	os.Setenv("MYPREFIX_MYDURATION", "22s")
	defer os.Unsetenv("MYPREFIX_MYDURATION")

	os.Setenv("MYPREFIX_MYMAP_MYKEY", "")
	defer os.Unsetenv("MYPREFIX_MYMAP_MYKEY")

	os.Setenv("MYPREFIX_MYMAP_MYKEY2_MYVALUE", "asd")
	defer os.Unsetenv("MYPREFIX_MYMAP_MYKEY2_MYVALUE")

	os.Setenv("MYPREFIX_MYMAP_MYKEY2_MYSTRUCT_MYPARAM", "456")
	defer os.Unsetenv("MYPREFIX_MYMAP_MYKEY2_MYSTRUCT_MYPARAM")

	os.Setenv("MYPREFIX_MYSLICE", "val1,val2")
	defer os.Unsetenv("MYPREFIX_MYSLICE")

	os.Setenv("MYPREFIX_MYSLICEEMPTY", "")
	defer os.Unsetenv("MYPREFIX_MYSLICEEMPTY")

	var s testStruct
	err := Load("MYPREFIX", &s)
	require.NoError(t, err)

	require.Equal(t, "testcontent", s.MyString)
	require.Equal(t, 123, s.MyInt)
	require.Equal(t, 15.2, s.MyFloat)
	require.Equal(t, true, s.MyBool)
	require.Equal(t, 22*myDuration(time.Second), s.MyDuration)

	_, ok := s.MyMap["mykey"]
	require.Equal(t, true, ok)

	v, ok := s.MyMap["mykey2"]
	require.Equal(t, true, ok)
	require.Equal(t, "asd", v.MyValue)
	require.Equal(t, 456, v.MyStruct.MyParam)

	require.Equal(t, []string{"val1", "val2"}, s.MySlice)
	require.Equal(t, []string{}, s.MySliceEmpty)
}
