package doc

// Tags groups the route under one or more named sections in Swagger UI.
//
//	doc.Doc[CreateUserRequest, UserResponse](docs, h.Create,
//	    doc.Summary("Register a new user"),
//	    doc.Tags("users"),
//	)
func Tags(tags ...string) DocOption {
	return func(op map[string]any) { op["tags"] = tags }
}
