package admin

import "time"

type Option func(*Admin)

func SetDurationTimebase(d time.Duration) Option {

	return func(a *Admin) {
		a.timeBase = d
	}
}
