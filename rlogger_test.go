package requestlog

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

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

func TestStatusLevels(t *testing.T) {
	for _, c := range []struct {
		code  int
		level string
	}{
		{http.StatusOK, "INFO"},
		{http.StatusNotModified, "INFO"},
		{http.StatusMovedPermanently, "INFO"},
		{http.StatusNotFound, "WARN"},
		{http.StatusForbidden, "ERRO"},
		{http.StatusInternalServerError, "ERRO"},
	} {
		req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.code)
		})
		b := bytes.Buffer{}
		New(h, Options{Logger: jplog.New(&b)}).ServeHTTP(w, req)
		out := stripAnsi(b.String())
		require.Contains(t, out, c.level, "status %d should log at %s, got: %s", c.code, c.level, out)
	}
}

// http.ResponseController only reaches the server through wrappers that
// implement Unwrap. Without it every control — notably the write deadline a
// handler needs to bound a slow client — returns http.ErrNotSupported.
func TestResponseControllerReachesTheServer(t *testing.T) {
	errc := make(chan error, 1)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errc <- http.NewResponseController(w).SetWriteDeadline(time.Now().Add(time.Minute))
		w.Write([]byte("bar"))
	})
	b := bytes.Buffer{}
	srv := httptest.NewServer(New(h, Options{Logger: jplog.New(&b)}))
	defer srv.Close()
	res, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer res.Body.Close()
	require.NoError(t, <-errc, "SetWriteDeadline must reach net/http through the monitorable writer")
}
