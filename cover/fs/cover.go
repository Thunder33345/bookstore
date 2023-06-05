package fs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/thunder33345/bookstore"
	"github.com/thunder33345/bookstore/http/rest"
)

type Store struct {
	//storeDir is the root directory where images are stored
	storeDir string
	//mountPoint is the web location of where Store.ShowCover is mounted
	mountPoint string
	db         dbStore
}

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
func (s *Store) StoreCover(ctx context.Context, bookID string, img io.ReadSeeker) error {
	//we detect and enforce the image types first
	ext, err := detectExt(img)
	if err != nil {
		return err
	}
	resourceName := bookID + ext

	//we create the img file, stored inside storeDir
	dst, err := os.Create(filepath.Join(s.storeDir, resourceName))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, img)
	if err != nil {
		return err
	}

	//finally we update the stored resource into our db
	return s.db.UpdateBookCover(ctx, bookID, resourceName)
}

// RemoveCover remove the stored cover file
func (s *Store) RemoveCover(ctx context.Context, bookID string) error {
	book, err := s.db.GetBook(ctx, bookID)
	if err != nil {
		return err
	}
	if book.CoverURL == "" {
		return nil
	}
	err = os.Remove(filepath.Join(s.storeDir, book.CoverURL))
	if err != nil {
		return err
	}
	return s.db.UpdateBookCover(ctx, bookID, "")
}

// ResolveCover turns a cover file into a URL
func (s *Store) ResolveCover(coverFile string) string {
	return s.mountPoint + coverFile
}

// HandleCoverRequest is a http handler mounted to match ResolveCover to display the cover file
func (s *Store) HandleCoverRequest(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "image")
	if fileName == "" {
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("no file name provided")))
		return
	}
	fmt.Printf("filename: %s\n", fileName)

	file, err := s.GetCoverFromResource(fileName)
	if err != nil {
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("failed opening file: %w", err)))
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("failed reading file")))
		return
	}
	typ, err := extToType(fileName)
	if err != nil {
		_ = render.Render(w, r, rest.ErrInvalidRequest(fmt.Errorf("failed getting file type")))
		return
	}
	w.Header().Set("Content-Type", typ)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf)

}

func (s *Store) GetCoverFromResource(resourceID string) (io.ReadCloser, error) {
	file, err := os.Open(s.getPath(resourceID))
	return file, err
}

func (s *Store) GetCover(ctx context.Context, bookId string) (io.ReadCloser, error) {
	book, err := s.db.GetBook(ctx, bookId)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(s.getPath(book.CoverURL))
	return file, err
}

func (s *Store) getPath(cover string) string {
	return filepath.Join(s.storeDir, cover)
}

type dbStore interface {
	GetBook(ctx context.Context, bookID string) (bookstore.Book, error)
	UpdateBookCover(ctx context.Context, bookID string, coverFile string) error
}

// detectExt detects file and provide relevant extension
// returns error if detected result is not jpeg or png
func detectExt(file io.ReadSeeker) (string, error) {
	//we create a buffer to detect img type
	//512 bytes because that's at most of what http.DetectContentType considers
	buff := make([]byte, 512)
	_, err := file.Read(buff)
	if err != nil {
		return "", err
	}

	ext, err := typeToExt(http.DetectContentType(buff))
	if err != nil {
		return "", err
	}

	//we seek back to the start before copying
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	return ext, nil
}

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
