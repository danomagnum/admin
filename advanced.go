package admin

type Notifyable interface {
	Changed()
}

type Changer interface {
	Change(val any)
}
