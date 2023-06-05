package rest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/moraes/isbn"
)

type ctxKey string

func (c ctxKey) String() string {
	return "bookstore context key: " + string(c)
}

var ctxKeyLimit = ctxKey("page-limit")
var ctxKeyAfter = ctxKey("page-after")

func (h *Handler) PaginationUUIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		limit := 50

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

		var afterID uuid.UUID
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

func (h *Handler) PaginationIBSNMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		limit := 50

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

		var afterID string
		if aft := q.Get("after"); aft != "" {
			afterID = aft
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxKeyAfter, afterID))

		next.ServeHTTP(w, r)
	})
}

var ctxUUIDKey = ctxKey("uuid")

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
