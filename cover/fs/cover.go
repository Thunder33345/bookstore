package fs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/thanhpk/randstr"
	"github.com/thunder33345/bookstore"
	"github.com/thunder33345/bookstore/http/rest"
)

// Store acts as an image store
// images are stored in file system, and the db maintains the filename
// this version provides HandleCoverRequest as a web mount to handle serving the files
type Store struct {
	//storeDir is the root directory where images are stored
	storeDir string
	//mountPoint is the web location of where Store.HandleCoverRequest is mounted
	mountPoint string
	//db allows image store to update book's cover metadata
	db dbStore
}

// NewStore creates a new image store
// webMount should describe where Store.HandleCoverRequest is mounted, this is necessary for generating canonical URL
// it should start with HTTP(s)://
func NewStore(fileDir string, webMount string, db dbStore) (*Store, error) {
	fileDir, err := filepath.Abs(fileDir)
	if err != nil {
		return nil, err
	}
	return &Store{
		storeDir:   fileDir,
		mountPoint: webMount,
		db:         db,
	}, nil
}

// StoreCover stores the cover file system
func (s *Store) StoreCover(ctx context.Context, isbn string, img io.ReadSeeker) error {
	//we detect and enforce the image types first
	fileType, err := detectType(img)
	if err != nil {
		return err
	}

	ext, err := typeToExt(fileType)
	if err != nil {
		return err
	}

	//a random padding helps with bypassing caching
	resourceName := isbn + "_" + randstr.Hex(4) + ext

	//we create the img file, stored inside storeDir
	dst, err := os.Create(s.getPath(resourceName))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, img)
	if err != nil {
		return err
	}

	//finally we update the stored resource into our db
	_, err = s.db.UpsertCoverData(ctx, bookstore.CoverData{
		ISBN:      isbn,
		CoverFile: resourceName,
	})
	return err
}

// RemoveCover remove the stored cover file from disk and db
func (s *Store) RemoveCover(ctx context.Context, isbn string) error {
	cover, err := s.db.GetCoverData(ctx, isbn)
	if err != nil {
		if isNoResultError(err) {
			return nil
		}
		return err
	}

	err = os.Remove(s.getPath(cover.CoverFile))
	//we allow removing from the db if the file no longer exist on disk
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = s.db.DeleteCoverData(ctx, isbn)
	if err != nil {
		return err
	}
	return nil
}

// GetCoverURL returns the cover URL if available, empty string is returned when there is no cover
func (s *Store) GetCoverURL(ctx context.Context, isbn string) (string, error) {
	data, err := s.db.GetCoverData(ctx, isbn)
	if err != nil {
		if isNoResultError(err) {
			return "", nil
		}
		return "", err
	}
	return s.mountPoint + data.CoverFile, nil
}

// ResolveCoverURL returns the cover URL from book data if available, empty string is returned when there is no cover
func (s *Store) ResolveCoverURL(_ context.Context, book bookstore.Book) (string, error) {
	if book.CoverData == nil || *book.CoverData == "" {
		return "", nil
	}
	return s.mountPoint + *book.CoverData, nil
}

// HandleCoverRequest is a http handler mounted to match ResolveCover to display the cover file
// it expects the {image} param to be available from chi
func (s *Store) HandleCoverRequest(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "image")
	if fileName == "" {
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("no file name provided")))
		return
	}

	file, err := os.Open(s.getPath(fileName))
	if err != nil {
		if os.IsNotExist(err) {
			_ = render.Render(w, r, rest.ErrNotFound)
			return
		}
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("failed opening file: %w", err)))
		return
	}
	defer file.Close()

	//we get the content type necessary for browsers to display it properly via file extension
	typ, err := extToType(fileName)
	if err != nil {
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("failed getting file type")))
		return
	}
	w.Header().Set("Content-Type", typ)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file)
}

// getPath is a helper to join create path prefixed with storeDir
func (s *Store) getPath(cover string) string {
	return filepath.Join(s.storeDir, cover)
}

// dbStore is a minimal interface of psql.Store
type dbStore interface {
	UpsertCoverData(ctx context.Context, cover bookstore.CoverData) (bookstore.CoverData, error)
	GetCoverData(ctx context.Context, isbn string) (bookstore.CoverData, error)
	DeleteCoverData(ctx context.Context, isbn string) error
}

func detectType(file io.ReadSeeker) (string, error) {
	//we create a buffer to detect img type
	//512 bytes because that's at most of what http.DetectContentType considers
	buff := make([]byte, 512)
	_, err := file.Read(buff)
	if err != nil {
		return "", err
	}

	fileType := http.DetectContentType(buff)

	//we seek back to the start before copying
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	return fileType, nil
}

// typeToExt convert filetype into an extension
func typeToExt(filetype string) (string, error) {
	switch filetype {
	case "image/jpeg":
		return ".jpeg", nil
	case "image/png":
		return ".png", nil
	default:
		return "", bookstore.ErrInvalidFileType
	}
}

// extToType convert file extension back into content type
func extToType(filename string) (string, error) {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpeg":
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	default:
		return "", fmt.Errorf("unknown file extension")
	}
}

func isNoResultError(err error) bool {
	var noRes *bookstore.NoResultError
	return errors.As(err, &noRes)
}
