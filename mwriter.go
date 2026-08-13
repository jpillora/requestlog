package requestlog

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/jpillora/sizestr"
	"github.com/tomasen/realip"
)

func (l *rlogger) monitorWriter(w http.ResponseWriter, r *http.Request) *monitorableWriter {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if l.Options.TrustProxy {
		ip = realip.FromRequest(r)
	}
	if ip == "127.0.0.1" || ip == "::1" {
		ip = "" //dont show localhost ips
	}
	return &monitorableWriter{
		l:      l,
		t0:     time.Now(),
		w:      w,
		r:      r,
		method: r.Method,
		path:   r.URL.Path,
		ip:     ip,
	}
}

// monitorable ResponseWriter
type monitorableWriter struct {
	l  *rlogger
	t0 time.Time
	//handler
	w http.ResponseWriter
	r *http.Request
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

// Unwrap exposes the wrapped ResponseWriter to http.ResponseController, so
// read/write deadlines and the rest of its controls reach the server through
// this layer. Without it every control returns http.ErrNotSupported, and a
// handler that relies on a write deadline to bound a slow client silently
// loses it.
func (m *monitorableWriter) Unwrap() http.ResponseWriter {
	return m.w
}

func (m *monitorableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := m.w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijacking not supported")
	}
	return hj.Hijack()
}

func (m *monitorableWriter) Flush() {
	m.w.(http.Flusher).Flush()
}

func (m *monitorableWriter) CloseNotify() <-chan bool {
	return m.w.(http.CloseNotifier).CloseNotify()
}

var integerRegexp = regexp.MustCompile(`\.\d+`)

// replace ResponseWriter with a monitorable one, return logger
func (m *monitorableWriter) done() {
	duration := time.Now().Sub(m.t0)
	if m.Code == 0 {
		m.Code = 200
	}
	if m.l.Options.Filter != nil && !m.l.Options.Filter(m.r, m.Code, duration, m.Size) {
		return //skip
	}
	args := []any{
		"method", m.method,
		"path", m.path,
		"code", m.Code,
	}
	if m.l.isJSON() {
		args = append(args, "duration", duration.Microseconds())
	} else {
		args = append(args, "duration", fmtDuration(duration))
	}
	if m.Size > 0 {
		if m.l.isJSON() {
			args = append(args, "size", m.Size)
		} else {
			args = append(args, "size", sizestr.ToString(m.Size))
		}
	}
	if m.ip != "" {
		args = append(args, "ip", m.ip)
	}
	log := m.levelStatus()
	log(http.StatusText(m.Code), args...)
}

type logFn func(msg string, args ...any)

func (m *monitorableWriter) levelStatus() logFn {
	switch {
	case m.Code == http.StatusNotFound:
		return m.l.Logger.Warn
	// 1xx/2xx/3xx are all normal outcomes. 3xx especially: a cache-revalidating
	// client turns every asset request into a 304, so levelling those as
	// warnings buries real ones under "Not Modified".
	case m.Code/100 == 1 || m.Code/100 == 2 || m.Code/100 == 3:
		return m.l.Logger.Info
	case m.Code/100 == 4 || m.Code/100 == 5:
		return m.l.Logger.Error
	}
	return m.l.Logger.Warn
}

var fmtDurationRe = regexp.MustCompile(`\.\d+`)

func fmtDuration(t time.Duration) string {
	return fmtDurationRe.ReplaceAllString(t.String(), "")
}
