package jsonpointer

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/dustin/gojson"
)

var ptests = []struct {
	path string
	exp  interface{}
}{
	{"/foo", []interface{}{"bar", "baz"}},
	{"/foo/0", "bar"},
	{"/", 0.0},
	{"/a~1b", 1.0},
	{"/c%d", 2.0},
	{"/e^f", 3.0},
	{"/g|h", 4.0},
	{"/i\\j", 5.0},
	{"/k\"l", 6.0},
	{"/ ", 7.0},
	{"/m~0n", 8.0},
	{"/g~1n~1r", "has slash, will travel"},
	{"/g/n/r", "where's tito?"},
}

func TestFindDecode(t *testing.T) {
	in := []byte(objSrc)

	var fl float64
	if err := FindDecode(in, "/g|h", &fl); err != nil {
		t.Errorf("Failed to decode /g|h: %v", err)
	}
	if fl != 4.0 {
		t.Errorf("Expected 4.0 at /g|h, got %v", fl)
	}

	fl = 0
	if err := FindDecode(in, "/z", &fl); err == nil {
		t.Errorf("Expected failure to decode /z: got %v", fl)
	}

	if err := FindDecode([]byte(`{"a": 1.x35}`), "/a", &fl); err == nil {
		t.Errorf("Expected failure, got %v", fl)
	}
}

func TestListPointers(t *testing.T) {
	got, err := ListPointers(nil)
	if err == nil {
		t.Errorf("Expected error on nil input, got %v", got)
	}
	got, err = ListPointers([]byte(`{"x": {"y"}}`))
	if err == nil {
		t.Errorf("Expected error on broken input, got %v", got)
	}
	got, err = ListPointers([]byte(objSrc))
	if err != nil {
		t.Fatalf("Error getting list of pointers: %v", err)
	}

	exp := []string{"", "/foo", "/foo/0", "/foo/1", "/", "/a~1b",
		"/c%d", "/e^f", "/g|h", "/i\\j", "/k\"l", "/ ", "/m~0n",
		"/g~1n~1r", "/g", "/g/n", "/g/n/r",
	}

	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, got)
	}
}

func TestPointerRoot(t *testing.T) {
	got, err := Find([]byte(objSrc), "")
	if err != nil {
		t.Fatalf("Error finding root: %v", err)
	}
	if !reflect.DeepEqual([]byte(objSrc), got) {
		t.Fatalf("Error finding root, found\n%s\n, wanted\n%s",
			got, objSrc)
	}
}

func TestPointerManyRoot(t *testing.T) {
	got, err := FindMany([]byte(objSrc), []string{""})
	if err != nil {
		t.Fatalf("Error finding root: %v", err)
	}
	if !reflect.DeepEqual([]byte(objSrc), got[""]) {
		t.Fatalf("Error finding root, found\n%s\n, wanted\n%s",
			got, objSrc)
	}
}

func TestPointerManyBroken(t *testing.T) {
	got, err := FindMany([]byte(`{"a": {"b": "something}}`), []string{"/a/b"})
	if err == nil {
		t.Errorf("Expected error parsing broken JSON, got %v", got)
	}
}

func TestPointerMissing(t *testing.T) {
	got, err := Find([]byte(objSrc), "/missing")
	if err != nil {
		t.Fatalf("Error finding missing item: %v", err)
	}
	if got != nil {
		t.Fatalf("Expected nil looking for /missing, got %v",
			got)
	}
}

func TestManyPointers(t *testing.T) {
	pointers := []string{}
	exp := map[string]interface{}{}
	for _, test := range ptests {
		pointers = append(pointers, test.path)
		exp[test.path] = test.exp
	}

	rv, err := FindMany([]byte(objSrc), pointers)
	if err != nil {
		t.Fatalf("Error finding many: %v", err)
	}

	got := map[string]interface{}{}
	for k, v := range rv {
		var val interface{}
		err = json.Unmarshal(v, &val)
		if err != nil {
			t.Fatalf("Error unmarshaling %s: %v", v, err)
		}
		got[k] = val
	}

	if !reflect.DeepEqual(got, exp) {
		for k, v := range exp {
			if !reflect.DeepEqual(got[k], v) {
				t.Errorf("At %v, expected %#v, got %#v", k, v, got[k])
			}
		}
		t.Fail()
	}
}

func TestManyPointersMissing(t *testing.T) {
	got, err := FindMany([]byte(objSrc), []string{"/missing"})
	if err != nil {
		t.Fatalf("Error finding missing item: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Expected empty looking for many /missing, got %v",
			got)
	}
}

var badDocs = [][]byte{
	[]byte{}, []byte(" "), nil,
	[]byte{'{'}, []byte{'['},
	[]byte{'}'}, []byte{']'},
}

func TestManyPointersBadDoc(t *testing.T) {
	for _, b := range badDocs {
		got, _ := FindMany(b, []string{"/broken"})
		if len(got) > 0 {
			t.Errorf("Expected failure on %v, got %v", b, got)
		}
	}
}

func TestPointersBadDoc(t *testing.T) {
	for _, b := range badDocs {
		got, _ := Find(b, "/broken")
		if len(got) > 0 {
			t.Errorf("Expected failure on %s, got %v", b, got)
		}
	}
}

func TestPointer(t *testing.T) {

	for _, test := range ptests {
		got, err := Find([]byte(objSrc), test.path)
		var val interface{}
		if err == nil {
			err = json.Unmarshal([]byte(got), &val)
		}
		if err != nil {
			t.Errorf("Got an error on key %v: %v", test.path, err)
			t.Fail()
		} else if !reflect.DeepEqual(val, test.exp) {
			t.Errorf("On %#v, expected %+v (%T), got %+v (%T)",
				test.path, test.exp, test.exp, val, val)
			t.Fail()
		} else {
			t.Logf("Success - got %s for %#v", got, test.path)
		}
	}
}

func TestPointerCoder(t *testing.T) {
	tests := map[string][]string{
		"/":        []string{""},
		"/a":       []string{"a"},
		"/a~1b":    []string{"a/b"},
		"/m~0n":    []string{"m~n"},
		"/ ":       []string{" "},
		"/g~1n~1r": []string{"g/n/r"},
		"/g/n/r":   []string{"g", "n", "r"},
	}

	for k, v := range tests {
		parsed := parsePointer(k)
		encoded := encodePointer(v)

		if k != encoded {
			t.Errorf("Expected to encode %#v as %#v, got %#v",
				v, k, encoded)
			t.Fail()
		}
		if !arreq(v, parsed) {
			t.Errorf("Expected to decode %#v as %#v, got %#v",
				k, v, parsed)
			t.Fail()
		}
	}
}

func TestCBugg406(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/pools.json")
	if err != nil {
		t.Fatalf("Error reading pools data: %v", err)
	}

	found, err := Find(data, "/implementationVersion")
	if err != nil {
		t.Fatalf("Failed to find thing: %v", err)
	}
	exp := ` "2.0.0-1976-rel-enterprise"`
	if string(found) != exp {
		t.Fatalf("Expected %q, got %q", exp, found)
	}
}

func BenchmarkEncodePointer(b *testing.B) {
	aPath := []string{"a", "ab", "a~0b", "a~1b", "a~0~1~0~1b"}
	for i := 0; i < b.N; i++ {
		encodePointer(aPath)
	}
}

func BenchmarkAll(b *testing.B) {
	obj := []byte(objSrc)
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Find(obj, test.path)
		}
	}
}

func BenchmarkManyPointer(b *testing.B) {
	pointers := []string{}
	for _, test := range ptests {
		pointers = append(pointers, test.path)
	}
	obj := []byte(objSrc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindMany(obj, pointers)
	}
}

func TestMustParseInt(t *testing.T) {
	tests := map[string]bool{
		"":   true,
		"0":  false,
		"13": false,
	}

	for in, out := range tests {
		var panicked bool
		func() {
			defer func() {
				panicked = recover() != nil
			}()
			mustParseInt(in)
			if panicked != out {
				t.Logf("Expected panicked=%v", panicked)
			}
		}()
	}
}

func TestFindBrokenJSON(t *testing.T) {
	x, err := Find([]byte(`{]`), "/foo/x")
	if err == nil {
		t.Errorf("Expected error, got %q", x)
	}
}

func TestGrokLiteral(t *testing.T) {
	brokenStr := "---broken---"
	tests := []struct {
		in  []byte
		exp string
	}{
		{[]byte(`"simple"`), "simple"},
		{[]byte(`"has\nnewline"`), "has\nnewline"},
		{[]byte(`"broken`), brokenStr},
	}

	for _, test := range tests {
		var got string
		func() {
			defer func() {
				if e := recover(); e != nil {
					got = brokenStr
				}
			}()
			got = grokLiteral(test.in)
		}()
		if test.exp != got {
			t.Errorf("Expected %q for %s, got %q",
				test.exp, test.in, got)
		}
	}
}

func TestSerieslySample(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/serieslysample.json")
	if err != nil {
		t.Fatalf("Error opening sample file: %v", err)
	}

	tests := []struct {
		pointer string
		exp     string
	}{
		{"/kind", "Listing"},
		{"/data/children/0/data/id", "w568e"},
		{"/data/children/0/data/name", "t3_w568e"},
	}

	for _, test := range tests {
		var found string
		err := FindDecode(data, test.pointer, &found)
		if err != nil {
			t.Errorf("Error on %v: %v", test.pointer, err)
		}
		if found != test.exp {
			t.Errorf("Expected %q, got %q", test.exp, found)
		}
	}
}

func TestSerieslySampleMany(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/serieslysample.json")
	if err != nil {
		t.Fatalf("Error opening sample file: %v", err)
	}

	keys := []string{"/kind", "/data/children/0/data/id", "/data/children/0/data/name"}
	exp := []string{` "Listing"`, ` "w568e"`, ` "t3_w568e"`}

	found, err := FindMany(data, keys)
	if err != nil {
		t.Fatalf("Error in FindMany: %v", err)
	}

	for i, k := range keys {
		if string(found[k]) != exp[i] {
			t.Errorf("Expected %q on %q, got %q", exp[i], k, found[k])
		}
	}
}

func TestSerieslySampleList(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/serieslysample.json")
	if err != nil {
		t.Fatalf("Error opening sample file: %v", err)
	}

	pointers, err := ListPointers(data)
	if err != nil {
		t.Fatalf("Error listing pointers: %v", err)
	}
	exp := 932
	if len(pointers) != exp {
		t.Fatalf("Expected %v pointers, got %v", exp, len(pointers))
	}
}

var codeJSON []byte

func init() {
	f, err := os.Open("testdata/code.json.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	codeJSON = data
}

func BenchmarkLarge3Key(b *testing.B) {
	keys := []string{
		"/tree/kids/0/kids/0/name",
		"/tree/kids/0/name",
		"/tree/kids/0/kids/0/kids/0/kids/0/kids/0/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 3 {
			b.Fatalf("Didn't find all the things from %v/%v",
				found, err)
		}
	}
}

func BenchmarkLargeShallow(b *testing.B) {
	keys := []string{
		"/tree/kids/0/kids/0/kids/1/kids/1/kids/3/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 1 {
			b.Fatalf("Didn't find all the things: %v/%v",
				found, err)
		}
	}
}

func BenchmarkLargeMissing(b *testing.B) {
	keys := []string{
		"/this/does/not/exist",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 0 {
			b.Fatalf("Didn't find all the things: %v/%v",
				found, err)
		}
	}
}

func BenchmarkLargeIdentity(b *testing.B) {
	keys := []string{
		"",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 1 {
			b.Fatalf("Didn't find all the things: %v/%v",
				found, err)
		}
	}
}

func BenchmarkLargeBest(b *testing.B) {
	keys := []string{
		"/tree/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 1 {
			b.Fatalf("Didn't find all the things: %v/%v",
				found, err)
		}
	}
}

func BenchmarkLargeMap(b *testing.B) {
	keys := []string{
		"/tree/kids/0/kids/0/kids/0/kids/0/kids/0/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		m := map[string]interface{}{}
		err := json.Unmarshal(codeJSON, &m)
		if err != nil {
			b.Fatalf("Error parsing JSON: %v", err)
		}
		Get(m, keys[0])
	}
}

const (
	tildeTestKey = "/name~0contained"
	slashTestKey = "/name~1contained"
)

func testDoubleReplacer(s string) string {
	return unescape(s)
}

func BenchmarkReplacerSlash(b *testing.B) {
	r := strings.NewReplacer("~1", "/", "~0", "~")
	for i := 0; i < b.N; i++ {
		r.Replace(slashTestKey)
	}
}

func BenchmarkReplacerTilde(b *testing.B) {
	r := strings.NewReplacer("~1", "/", "~0", "~")
	for i := 0; i < b.N; i++ {
		r.Replace(tildeTestKey)
	}
}

func BenchmarkDblReplacerSlash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testDoubleReplacer(slashTestKey)
	}
}

func BenchmarkDblReplacerTilde(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testDoubleReplacer(tildeTestKey)
	}
}
