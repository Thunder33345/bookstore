package rest

import (
	"net/http"

	"github.com/go-chi/render"
)

func (h *Handler) UpdateBookCover(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxISBNKey).(string)

	//limit max file size to 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequestBody(err))
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		_ = render.Render(w, r, ErrProcessingFile(err))
		return
	}
	defer file.Close()

	err = h.cover.StoreCover(r.Context(), id, file)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}
	render.Status(r, http.StatusNoContent)
}

func (h *Handler) DeleteBookCover(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxISBNKey).(string)
	err := h.cover.RemoveCover(r.Context(), id)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}
	render.Status(r, http.StatusNoContent)
}
