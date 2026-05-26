package apidoc

// DocOption applies metadata directly to an OpenAPI operation map.
// Each option is self-contained and knows exactly which field to set,
// so docMeta never needs a new field when a new option is added (OCP).
type DocOption func(op map[string]any)
