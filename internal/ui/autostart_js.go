//go:build js && wasm

package ui

import (
	"strconv"
	"strings"
	"syscall/js"
)

// autostartROM reads ?rom=N from the page URL to deep-link a bundled game.
func autostartROM() int {
	search := js.Global().Get("location").Get("search").String()
	for q := range strings.SplitSeq(strings.TrimPrefix(search, "?"), "&") {
		if v, ok := strings.CutPrefix(q, "rom="); ok {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				return n
			}
		}
	}
	return -1
}
