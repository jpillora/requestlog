package requestlog

import (
	"regexp"
	"time"

	"github.com/jpillora/ansi"
)

type Colors struct {
	Grey, Cyan, Yellow, Red, Reset string
}

var defaultColors = &Colors{string(ansi.Set(ansi.Black)), string(ansi.Set(ansi.Cyan)), string(ansi.Set(ansi.Yellow)), string(ansi.Set(ansi.Yellow)), string(ansi.Set(ansi.Reset))}
var noColors = &Colors{} //no colors

func colorcode(status int) string {
	switch status / 100 {
	case 2:
		return string(ansi.Set(ansi.Green))
	case 3:
		return string(ansi.Set(ansi.Cyan))
	case 4:
		return string(ansi.Set(ansi.Yellow))
	case 5:
		return string(ansi.Set(ansi.Red))
	}
	return string(ansi.Set(ansi.Black))
}

var fmtdurationRe = regexp.MustCompile(`\.\d+`)

func fmtduration(t time.Duration) string {
	return fmtdurationRe.ReplaceAllString(t.String(), "")
}
