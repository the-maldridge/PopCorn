package stats

var (
	callbacks []Callback
)

// RegisterCallback registers a callback to be executed at a later
// time.
func RegisterCallback(c Callback) {
	callbacks = append(callbacks, c)
}

// DoCallbacks processes all callbacks serially.
func DoCallbacks() {
	for _, c := range callbacks {
		c()
	}
}
