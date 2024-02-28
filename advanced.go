package admin

// Any struct implementing the Notifyable interface will have the Changed() function called after it is modified.
// This will be called whether or not all the values are identical to their previous values.
type Notifyable interface {
	Changed()
}

// Any struct implementing the Changer interface will have the Change() function called instead of being modified
// by the admin page itself.  It is the responsibility of the struct to update its fields.
//
// val will be a pointer to a new instance of the structure with the returned values.  This will have to be asserted
// by the function implementation itself.
//
// After change is called, the Notifyable interface is checked and called regardless of what this function does.
type Changer interface {
	Change(val any)
}

// Any struct implementing the Deleteable interface will have a [delete] button show up when editing the data.
// If clicked, the struct will be unregistered from the admin page and the Delete() function will be called.
// If the delete is abortable, the Delete() function will have to Register the struct again.
type Deleteable interface {
	Delete()
}
