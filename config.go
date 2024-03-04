package admin

import "time"

func SetDurationTimebase(d time.Duration) func(*Admin) {

	return func(a *Admin) {
		a.timeBase = d
	}
}
