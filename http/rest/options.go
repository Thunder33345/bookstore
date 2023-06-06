package rest

// Option is a callable that modifies the Handler's parameter
type Option func(h Handler) Handler

// WithMaxListing changes the default max element listing
func WithMaxListing(limit int) Option {
	return func(h Handler) Handler {
		h.maxListLimit = limit
		return h
	}
}

// WithIgnoreInvalidISBN disable ISBN validation when creating books
func WithIgnoreInvalidISBN(b bool) Option {
	return func(h Handler) Handler {
		h.ignoreInvalidIBSN = b
		return h
	}
}
