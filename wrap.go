package requestlog

import (
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/andrew-d/go-termutil"
	"github.com/jpillora/sizestr"
)

type Options struct {
	Writer     io.Writer
	TimeFormat string
	Format     string
	formatTmpl *template.Template
	Colors     *Colors
}

var defaultOptions = Options{
	Writer:     os.Stdout,
	TimeFormat: "2006/01/02 15:04:05.000",
	Format: `{{ .Grey }}{{ if .Timestamp }}{{ .Timestamp }} {{end}}` +
		`{{ .Method }} {{ .Path }} {{ .CodeColor }}{{ .Code }}{{ .Grey }} ` +
		`{{ .Duration }}{{ if .Size }} {{ .Size }}{{end}}` +
		`{{ if .IP }} ({{ .IP }}){{end}}{{ .Reset }}` + "\n",
	Colors: nil,
}

var isInteractive = termutil.Isatty(os.Stdout.Fd())

func init() {
	sizestr.ToggleCase()
	if isInteractive {
		defaultOptions.Colors = defaultColors
	} else {
		defaultOptions.Colors = noColors
	}
}

func Wrap(next http.Handler) http.Handler {
	return WrapWith(next, &defaultOptions)
}

func WrapWith(next http.Handler, opts *Options) http.Handler {
	var err error
	opts.formatTmpl, err = template.New("format").Parse(opts.Format)
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := monitorWriter(w, r, opts)
		next.ServeHTTP(m, r)
		m.Log()
	})
}
