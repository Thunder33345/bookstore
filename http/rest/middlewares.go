package rest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/moraes/isbn"
	"github.com/thunder33345/bookstore"
)

// ctxKey is an unexported type to prevent context key collisions
type ctxKey string

func (c ctxKey) String() string {
	return "bookstore context key: " + string(c)
}

var ctxKeyLimit = ctxKey("page-limit")

// PaginationLimitMiddleware populates the ctxKeyLimit for controlling listed elements in pagination
func (h *Handler) PaginationLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		limit := h.defaultListLimit

		var err error
		if lim := q.Get("limit"); lim != "" {
			limit, err = strconv.Atoi(lim)
			if err != nil {
				_ = render.Render(w, r, ErrInvalidRequestParam("limit", err))
				return
			}
			if limit > h.maxListLimit {
				_ = render.Render(w, r, ErrInvalidRequestParam("limit", fmt.Errorf("invalid value %d", limit)))
				return
			}
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxKeyLimit, limit))
		next.ServeHTTP(w, r)
	})
}

var ctxKeyAfter = ctxKey("page-after")

// PaginationUUIDMiddleware populates the ctxKeyAfter with a UUID
func (h *Handler) PaginationUUIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var afterID uuid.UUID
		var err error
		if aft := q.Get("after"); aft != "" {
			afterID, err = uuid.Parse(aft)
			if err != nil {
				_ = render.Render(w, r, ErrInvalidRequestParam("after", err))
				return
			}
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxKeyAfter, afterID))
		next.ServeHTTP(w, r)
	})
}

// PaginationIBSNMiddleware populates the ctxKeyAfter with a string
func (h *Handler) PaginationIBSNMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var afterID string
		if aft := q.Get("after"); aft != "" {
			afterID = aft
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxKeyAfter, afterID))
		next.ServeHTTP(w, r)
	})
}

var ctxUUIDKey = ctxKey("uuid")

// UUIDCtx populates the UUID into context from url param, and perform validation
func UUIDCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid uuid.UUID
		var err error
		id := chi.URLParam(r, "uuid")
		if id == "" {
			_ = render.Render(w, r, ErrInvalidIDRequest(fmt.Errorf("UUID not provided")))
			return
		}
		uid, err = uuid.Parse(id)
		if err != nil || uid == uuid.Nil {
			_ = render.Render(w, r, ErrInvalidIDRequest(err))
			return
		}

		ctx := context.WithValue(r.Context(), ctxUUIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var ctxISBNKey = ctxKey("isbn")

// ISBNCtx populates the ISBN into context from url param
func ISBNCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id string
		//var err error
		id = chi.URLParam(r, "isbn")
		if id == "" {
			_ = render.Render(w, r, ErrInvalidIDRequest(fmt.Errorf("ISBN not provided")))
			return
		}

		//we assume anything 13 digits is valid ISBN
		if idLen := len(id); idLen != 13 {
			_ = render.Render(w, r, ErrInvalidIDRequest(fmt.Errorf("invalid ISBN provided")))
			return
		}

		ctx := context.WithValue(r.Context(), ctxISBNKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MiddlewareAuthenticatedOnly is a middleware to enforce authentication
func (h *Handler) MiddlewareAuthenticatedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//try to populate the session into context
		//so other handlers can also use them
		r, _, err := h.populateSession(r)
		if err != nil {
			_ = render.Render(w, r, ErrSessionResponse(err))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// MiddlewareAdminOnly is a middleware to enforce admin only
func (h *Handler) MiddlewareAdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//try to populate the session into context
		//so other handlers can also use them
		r, account, err := h.populateSession(r)
		if err != nil {
			_ = render.Render(w, r, ErrSessionResponse(err))
			return
		}
		if account.Admin == false {
			_ = render.Render(w, r, ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// populateSession tries to populate session data into context using header
func (h *Handler) populateSession(r *http.Request) (*http.Request, bookstore.Session, error) {
	//if it's already populated, we skip it
	//this could happen if middlewares got chained
	if ses, ok := GetSession(r.Context()); ok {
		return r, ses, nil
	}

	ah := r.Header.Get("Authorization")
	if ah == "" {
		return r, bookstore.Session{}, bookstore.ErrMissingSession
	}
	if !strings.HasPrefix(ah, "Bearer ") {
		return r, bookstore.Session{}, bookstore.ErrMalformedSession
	}
	ah = strings.TrimLeft(ah, "Bearer ")

	account, err := h.auth.GetSession(r.Context(), ah)
	if err != nil {
		return r, bookstore.Session{}, err
	}
	r = r.WithContext(context.WithValue(r.Context(), ctxKey("user"), account))
	r = r.WithContext(context.WithValue(r.Context(), ctxKey("token"), ah))
	return r, account, nil
}

func GetSession(ctx context.Context) (bookstore.Session, bool) {
	ses, ok := ctx.Value(ctxKey("user")).(bookstore.Session)
	return ses, ok
}

func GetSessionToken(ctx context.Context) (string, bool) {
	tok, ok := ctx.Value(ctxKey("token")).(string)
	return tok, ok
}

func GetSessionAndToken(ctx context.Context) (bookstore.Session, string, bool) {
	ses, okU := ctx.Value(ctxKey("user")).(bookstore.Session)
	tok, okT := ctx.Value(ctxKey("token")).(string)
	return ses, tok, okU && okT
}

// validateISBN validates ISBN upon book creation
// h.ignoreInvalidIBSN allows ignoring validating ISBN checksum
func (h *Handler) validateISBN(id string) (string, error) {
	var err error
	if idLen := len(id); idLen != 10 && idLen != 13 {
		return "", fmt.Errorf("invalid ISBN provided(length=%d), should be 10 or 13", idLen)
	} else if idLen == 10 {
		id, err = isbn.To13(id)
		if err != nil {
			return "", fmt.Errorf("error converting to ISBN-13: %w", err)
		}
	}
	if h.ignoreInvalidIBSN {
		return id, nil
	}

	if !isbn.Validate13(id) {
		return "", fmt.Errorf("invalid ISBN provided")
	}
	return id, err
}

// stringSliceToUUID is a utility function to convert a slice of string into a slice of uuid
func stringSliceToUUID(ids []string) ([]uuid.UUID, error) {
	if ids == nil || len(ids) == 0 {
		return nil, nil
	}
	uidList := make([]uuid.UUID, 0, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("error parsing #%d argument: %w", i, err)
		}
		uidList = append(uidList, uid)
	}
	return uidList, nil
}
