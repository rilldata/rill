package ai

import "time"

// PreloadSkills exposes preloadSkills to the external test package.
var PreloadSkills = preloadSkills

// SetClickUITimeout overrides how long click_ui waits for the browser, restoring it when the test ends.
func SetClickUITimeout(t interface{ Cleanup(func()) }, d time.Duration) {
	prev := clickUITimeout
	clickUITimeout = d
	t.Cleanup(func() { clickUITimeout = prev })
}
