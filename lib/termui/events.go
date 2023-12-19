package termui

import "github.com/qbradq/after/lib/util"

type EventKey struct {
	Key rune
}

type EventResize struct {
	Size util.Point
}
