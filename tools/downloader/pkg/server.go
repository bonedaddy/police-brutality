package pkg

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/multierr"
)

// Server is a github webhook server to enable auto-regeneration of
type Server struct {
	srv     *http.Server
	dl      *Downloader
	up      Uploader
	tlsCert string
	tlsKey  string
}

// ServerOpts is used to confiugre the webhook server
type ServerOpts struct {
	ListenAddress string
	TLSCert       string
	TLSKey        string
}

// NewServer returns an initialized webhook server
func NewServer(opts ServerOpts, dl *Downloader, up Uploader) *Server {
	router := chi.NewRouter()
	router.HandleFunc("/github/webhook/payload", func(w http.ResponseWriter, r *http.Request) {
		var out map[interface{}]interface{}
		if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
			handleErr(err, w)
			return
		}
		log.Println("new payload received: ", out)
	})
	return &Server{
		srv:     &http.Server{Addr: opts.ListenAddress, Handler: router},
		dl:      dl,
		up:      up,
		tlsKey:  opts.TLSKey,
		tlsCert: opts.TLSCert,
	}
}

// Run starts the webhook server
func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		if s.tlsCert != "" && s.tlsKey != "" {
			errChan <- s.srv.ListenAndServeTLS(s.tlsCert, s.tlsKey)
			return
		}
		errChan <- s.srv.ListenAndServe()
	}()
	<-ctx.Done()
	err := s.srv.Close()
	return multierr.Combine(err, <-errChan)
}

func handleErr(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
