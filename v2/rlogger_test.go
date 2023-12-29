package requestlog_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jpillora/jplog"
	"github.com/jpillora/requestlog/v2"
)

func TestOut(t *testing.T) {

	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})

	b := bytes.Buffer{}

	wrap := requestlog.Wrap(h, requestlog.Options{
		Logger: jplog.New(&b),
	})

	wrap.ServeHTTP(w, req)

	t.Log(b.String())
}
