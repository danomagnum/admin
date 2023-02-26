package gowebstructapi

type schemaEndpoint struct {
	Type    string // can be object, array, string, integer
	Format  string // can be choices, ?
	Enum    []string
	Options schemaOptions
}
type schemaOptions struct {
	enum_titles []string
}
