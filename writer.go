package requestlog

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/jpillora/sizestr"
)

func monitorWriter(w http.ResponseWriter, r *http.Request, opts *Options) *monitorableWriter {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	if ip == "127.0.0.1" || ip == "::1" {
		ip = ""
	}
	return &monitorableWriter{
		opts:   opts,
		t0:     time.Now(),
		w:      w,
		method: r.Method,
		path:   r.URL.Path,
		ip:     ip,
	}
}

//monitorable ResponseWriter
type monitorableWriter struct {
	opts *Options
	t0   time.Time
	//handler
	w http.ResponseWriter
	//stats
	method, path, ip string
	Code             int
	Size             int64
}

func (m *monitorableWriter) Header() http.Header {
	return m.w.Header()
}

func (m *monitorableWriter) Write(p []byte) (int, error) {
	m.Size += int64(len(p))
	return m.w.Write(p)
}

func (m *monitorableWriter) WriteHeader(c int) {
	m.Code = c
	m.w.WriteHeader(c)
}

var integerRegexp = regexp.MustCompile(`\.\d+`)

//replace ResponseWriter with a monitorable one, return logger
func (m *monitorableWriter) Log() {
	now := time.Now()
	if m.Code == 0 {
		m.Code = 200
	}
	cc := ""
	if isInteractive {
		cc = colorcode(m.Code)
	}
	size := ""
	if m.Size > 0 {
		size = sizestr.ToString(m.Size)
	}
	buff := bytes.Buffer{}
	m.opts.formatTmpl.Execute(&buff, &struct {
		*Colors
		Timestamp, Method, Path, CodeColor string
		Code                               int
		Duration, Size, IP                 string
	}{
		m.opts.Colors,
		m.t0.Format(m.opts.TimeFormat), m.method, m.path, cc,
		m.Code,
		fmtduration(now.Sub(m.t0)), size, m.ip,
	})
	//fmt is threadsafe :)
	fmt.Fprint(m.opts.Writer, buff.String())
}
