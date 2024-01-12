//go:build after_debug

package util

import "log"

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func Log(fmt string, args ...any) {
	log.Printf(fmt, args...)
}
