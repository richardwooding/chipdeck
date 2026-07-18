//go:build !js

package ui

// autostartROM is a no-op outside the browser.
func autostartROM() int { return -1 }
