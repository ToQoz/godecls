package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestProccessFile(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{
			Input: `package main
import "fmt"

var (
	foo = "bar"
)

func main() {
}
			`,
			Expected: `var foo = "bar"
func main() {...}
`,
		},
		{
			Input: `package main
import "fmt"

var (
	foo = "bar"
	bar = "foo"
)

var (
	foobar = "barfoo"
)

const (
	baz = "baz"
)

func main() {
}

var f = func(a string) (b string, c error) {
}
			`,
			Expected: `var foo = "bar"
var bar = "foo"
var foobar = "barfoo"
const baz = "baz"
func main() {...}
var f = func(a string) (b string, c error) {...}
`,
		},
	}

	for _, test := range tests {
		var out bytes.Buffer
		in := bytes.NewReader([]byte(test.Input))

		err := proccessFile("<no filename>", &out, in)
		if err != nil {
			t.Fatal(err.Error())
		}

		got := out.String()
		if got != test.Expected {
			errStr := fmt.Sprintf("proccessFile outputs unexpected string\n# expected\n%s\n\n# got\n%s\n\n", quote(test.Expected), quote(got))
			if diff, err := diffString(test.Expected, got); err == nil {
				errStr += diff
			}
			t.Error(errStr)
		}
	}

}

func diffString(a, b string) (string, error) {
	diff, err := diffBytes([]byte(a), []byte(b))
	if err != nil {
		return "", err
	}
	return string(diff), nil
}

func diffBytes(a, b []byte) ([]byte, error) {
	f1, err := ioutil.TempFile("", "gofmt")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "gofmt")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(a)
	f2.Write(b)

	return cmdDiff(f1.Name(), f2.Name())
}

func cmdDiff(f1, f2 string) ([]byte, error) {
	d, err := exec.Command("diff", "-u", f1, f2).CombinedOutput()
	if err != nil && len(d) == 0 {
		return nil, err
	}

	return d, nil
}

func quote(a string) string {
	return "`" + a + "`"
}
