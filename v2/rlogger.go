package requestlog

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jpillora/jplog"
)

type Options struct {
	Group      string
	Logger     *slog.Logger
	Filter     func(r *http.Request, code int, duration time.Duration, size int64) bool
	TrustProxy bool //TrustProxy will log X-Forwarded-For/X-Real-Ip instead of the IP source
}

func Wrap(next http.Handler, opts ...Options) http.Handler {
	o := Options{}
	if len(opts) > 0 {
		o = opts[0]
	}
	return newRequestLogger(next, o)
}

type rlogger struct {
	next http.Handler
	Options
	*slog.Logger
}

func (l *rlogger) isJSON() bool {
	_, ok := l.Logger.Handler().(*slog.JSONHandler)
	return ok
}

func newRequestLogger(next http.Handler, opts Options) *rlogger {
	l := opts.Logger
	g := opts.Group
	if l == nil && g == "" {
		g = "http"
	}
	if l == nil {
		l = jplog.New(os.Stdout)
	}
	if g != "" {
		l = l.WithGroup(g)
	}
	return &rlogger{
		next,
		opts,
		l,
	}
}

// serve http
func (l *rlogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := l.monitorWriter(w, r)
	l.next.ServeHTTP(m, r)
	m.done()
}
