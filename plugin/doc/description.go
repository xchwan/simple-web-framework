package doc

// Description sets a long-form description for the route, shown when the endpoint is expanded in Swagger UI.
//
//	doc.Doc[CreateUserRequest, UserResponse](docs, h.Create,
//	    doc.Summary("Register a new user"),
//	    doc.Description("Creates a new user account. The email address must be unique."),
//	)
func Description(s string) DocOption {
	return func(m *docMeta) { m.description = s }
}
