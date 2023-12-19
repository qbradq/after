package main

import (
	"github.com/qbradq/after/internal/client/termgui"
	"github.com/qbradq/after/lib/tcelldriver"
)

func main() {
	termgui.Main(&tcelldriver.Driver{})
}
