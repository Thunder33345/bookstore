package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
	"github.com/thunder33345/bookstore/auth"
	"github.com/thunder33345/bookstore/cover/fs"
	"github.com/thunder33345/bookstore/db/psql"
	"github.com/thunder33345/bookstore/http/rest"
)

var routes = flag.Bool("routes", false, "Generate router documentation")
var debugRoutes = flag.Bool("debug-routes", false, "Mount unprotected debug route")
var debugIgnoreInvalidISBN = flag.Bool("debug-isbn", false, "Disable ISBN validation")

func main() {
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		fmt.Printf("Continuing anyways...\n")
	}

	if os.Getenv("DATABASE_URL") == "" {
		fmt.Printf("ENV DATABASE_URL missing\nShould be the connection string used to connect to psql DB.")
		return
	}

	if os.Getenv("URL") == "" {
		fmt.Printf("ENV URL missing\nShould be the canonical URL.")
		return
	}

	if os.Getenv("LISTEN") == "" {
		fmt.Printf("ENV LISTEN missing\nShould be an address to listen on(0.0.0.0:8080).")
	}

	fmt.Printf("Initilizing db\n")
	db, err := psql.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	err = db.Init()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Initilizing auth service\n")
	authService := auth.NewAuth(db, 10)

	fmt.Printf("Initilizing cover store\n")
	coverService, err := fs.NewStore("./data/covers", os.Getenv("URL")+"/covers/", db)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Initilizing REST handler\n")
	restService := rest.NewHandler(db, coverService, authService, rest.WithIgnoreInvalidISBN(*debugIgnoreInvalidISBN))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("root"))
	})

	//mount the API
	r.Route("/api/v1/", func(r chi.Router) {
		r.Use(middleware.URLFormat)
		r.Use(render.SetContentType(render.ContentTypeJSON))
		restService.Mount(r)

		if *debugRoutes {
			//intentionally exposes the user management endpoint without any auth middleware
			r.Route("/debug/users", func(r chi.Router) {
				fmt.Printf("Mounting unprotected debug router: /api/v1/debug/users")
				r.With(restService.PaginationLimitMiddleware, restService.PaginationUUIDMiddleware).Get("/", restService.ListUsers)
				r.Post("/", restService.CreateUser)
				r.With(rest.UUIDCtx).Route("/{uuid}", func(r chi.Router) {
					r.Get("/", restService.GetUser)
					r.Put("/", restService.UpdateUser)
					r.Delete("/", restService.DeleteUser)
					r.Post("/password", restService.UpdateUserPassword)
					r.Delete("/password", restService.DeleteUserSessions)
					r.Delete("/sessions", restService.DeleteUserSessions)
				})
			})
		}
	})

	//mount the cover handler
	r.Get("/covers/{image}", coverService.HandleCoverRequest)

	fmt.Printf("Initilizing server\n")
	server := &http.Server{Addr: os.Getenv("LISTEN"), Handler: r}

	//some context and signals for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				fmt.Printf("graceful shutdown timed out.. forcing exit.\n")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			fmt.Printf("error while shutting down: %v\n", err)
		}
		serverStopCtx()
	}()

	if *routes {
		fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
			ProjectPath: "github.com/thunder33345/bookstore",
		}))
		return
	}

	fmt.Printf("Listening for request on %s\n", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Error while listening: %v\n", err)
	}

	<-serverCtx.Done()
	fmt.Printf("Server Exited\n")
}
