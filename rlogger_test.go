package requestlog

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/jpillora/jplog"
	"github.com/stretchr/testify/require"
)

func TestOut(t *testing.T) {

	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})

	b := bytes.Buffer{}

	wrap := New(h, Options{
		Logger: jplog.New(&b),
	})

	wrap.ServeHTTP(w, req)

	out := stripAnsi(b.String())
	require.Contains(t, out, `INFO OK method=GET path=/foo code=200`)
}

var ansire = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")

func stripAnsi(str string) string {
	return ansire.ReplaceAllString(str, "")
}
