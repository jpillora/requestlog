package requestlog

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jpillora/sizestr"
)

var Writer = io.Writer(os.Stdout)
var TimeFormat = "2006/01/02 15:04:05.000"

func Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := New(w)
		next.ServeHTTP(m, r)
		m.Log(r)
	})
}

func New(w http.ResponseWriter) *MonitorableWriter {
	return &MonitorableWriter{
		t0: time.Now(),
		w:  w,
	}
}

//monitorable ResponseWriter
type MonitorableWriter struct {
	t0 time.Time
	//
	w http.ResponseWriter
	//stats
	Code int
	Size int64
}

func (m *MonitorableWriter) Header() http.Header {
	return m.w.Header()
}

func (m *MonitorableWriter) Write(p []byte) (int, error) {
	m.Size += int64(len(p))
	return m.w.Write(p)
}

func (m *MonitorableWriter) WriteHeader(c int) {
	m.Code = c
	m.w.WriteHeader(c)
}

var integerRegexp = regexp.MustCompile(`\.\d+`)

//replace ResponseWriter with a monitorable one, return logger
func (m *MonitorableWriter) Log(r *http.Request) {

	now := time.Now()

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	if m.Code == 0 {
		m.Code = 200
	}
	size := sizestr.ToString(m.Size)

	b := bytes.Buffer{}
	b.WriteString(now.Format(TimeFormat) + " ")
	b.WriteString(r.Method + " ")
	b.WriteString(r.URL.Path + " ")
	b.WriteString(strconv.Itoa(m.Code) + " ")
	dur := integerRegexp.ReplaceAllString(now.Sub(m.t0).String(), "")
	b.WriteString(dur + " ")
	if size != "0B" {
		b.WriteString(size + " ")
	}
	if ip != "::1" && ip != "127.0.0.1" {
		b.WriteString("(" + ip + ")")
	}
	b.WriteString("\n")
	Writer.Write(b.Bytes())
}