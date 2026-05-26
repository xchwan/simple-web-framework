package doc

// Summary sets a short description for the route, displayed as the endpoint title in Swagger UI.
//
//	doc.Doc[CreateUserRequest, UserResponse](docs, h.Create, doc.Summary("Register a new user"))
//
// A plain string is also accepted as a shorthand:
//
//	doc.Doc[CreateUserRequest, UserResponse](docs, h.Create, "Register a new user")
func Summary(s string) DocOption {
	return func(m *docMeta) { m.summary = s }
}
