package nfc

import "log"

var debug bool

// SetDebug enables or disables debug logging.
func SetDebug(on bool) { debug = on }

// Debugf logs when debug is enabled.
func Debugf(format string, args ...any) {
	if debug {
		log.Printf("[gonfc] "+format, args...)
	}
}
