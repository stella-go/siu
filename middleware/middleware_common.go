package middleware

type f func(format string, arr ...interface{})

// adjusts call stack depth offset between p.Logger and siu loggers to ensure accurate source code line attribution
func printLogger(f f, format string, arr ...interface{}) {
	func() {
		f(format, arr...)
	}()
}
