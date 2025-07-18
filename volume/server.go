package volume

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	port   string
	volume *Volume
	volsrv *http.Server
}

func NewServer(port string, v *Volume) *Server {
	srv := &Server{
		volume: v,
	}
	router := chi.NewRouter()
	router.Post("/", srv.handlePut)
	router.Get("/{uuid}", srv.handleGet)

	srv.volsrv = &http.Server{
		Addr:    port,
		Handler: router,
	}
	return srv
}

func (srv *Server) Start() {
	err := srv.volsrv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}

func (srv *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	req, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		jsonResponse(
			w, http.StatusInternalServerError,
			map[string]string{"error": "Couldnt read req body"},
		)
		return
	}

	objectId, err := srv.volume.Write(req)
	if err != nil {
		jsonResponse(
			w, http.StatusInternalServerError,
			map[string]string{"error": "Unable to write data"},
		)
		return
	}
	resp := map[string]string{
		"message":   "Successfully wrote data to new needle",
		"object_id": objectId.String(),
	}
	jsonResponse(w, http.StatusCreated, resp)
}

func (srv *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	objectIdStr := chi.URLParam(r, "uuid")
	if len(objectIdStr) == 0 {
		jsonResponse(
			w, http.StatusBadRequest,
			map[string]string{"error": "No object ID provided"},
		)
		return
	}

	objectId, err := uuid.Parse(objectIdStr)
	if err != nil {
		jsonResponse(
			w, http.StatusInternalServerError,
			map[string]string{"error": "Couldn't parse object id into byte arr"},
		)
		return
	}
	data, err := srv.volume.Read(objectId)
	if err != nil {
		var ErrNotFound = errors.New("object not found")
		if errors.Is(err, ErrNotFound) {
			jsonResponse(
				w, http.StatusNotFound,
				map[string]string{"error": "object not found"},
			)
		} else {
			jsonResponse(
				w, http.StatusInternalServerError,
				map[string]string{"error": "failed to read object"},
			)
		}
		return
	}

	octetStreamResponse(w, http.StatusOK, data)
}

func octetStreamResponse(w http.ResponseWriter, status int, data []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	_, err := w.Write(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error" : "failed to write binary to response"}`))
		return errors.New("failed to write binary to response")
	}
	return nil
}

func jsonResponse(w http.ResponseWriter, status int, respPayload any) error {
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(respPayload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error" : "failed to create json resp"}`))
		return errors.New("Couldn't marshal json")
	}
	w.WriteHeader(status)
	w.Write(j)
	return nil
}
