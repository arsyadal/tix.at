package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"time"

	"tix.at/internal/events"
	"tix.at/internal/rabbit"
	"tix.at/internal/stock"
	"tix.at/internal/store"
)

type Server struct {
	store      *store.Store
	stock      *stock.Store
	mq         *rabbit.Client
	bookingTTL time.Duration
}

func New(store *store.Store, stock *stock.Store, mq *rabbit.Client, bookingTTL time.Duration) *Server {
	return &Server{store: store, stock: stock, mq: mq, bookingTTL: bookingTTL}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /events", s.createEvent)
	mux.HandleFunc("GET /events", s.listEvents)
	mux.HandleFunc("POST /bookings", s.createBooking)

	mux.Handle("GET /{$}", http.FileServer(http.Dir("public")))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("public"))))
	return mux
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.GetEvents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

type eventReq struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	var req eventReq
	if err := readJSON(r, &req); err != nil || req.ID == "" || req.Name == "" || req.Stock < 0 {
		writeError(w, http.StatusBadRequest, "invalid event")
		return
	}
	if err := s.store.UpsertEvent(r.Context(), req.ID, req.Name, req.Stock); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.stock.Set(r.Context(), req.ID, req.Stock); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

type bookingReq struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	var req bookingReq
	if err := readJSON(r, &req); err != nil || req.EventID == "" || req.UserID == "" {
		writeError(w, http.StatusBadRequest, "invalid booking")
		return
	}
	ok, err := s.stock.Reserve(r.Context(), req.EventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusConflict, "sold out")
		return
	}

	bookingID := randID()
	if err := s.store.CreateBooking(r.Context(), bookingID, req.EventID, req.UserID, s.bookingTTL); err != nil {
		_ = s.stock.Release(context.Background(), req.EventID)
		if s.store.IsUniqueViolation(err) {
			writeError(w, http.StatusConflict, "already booked")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	msg := events.BookingCreated{BookingID: bookingID, EventID: req.EventID, UserID: req.UserID}
	if err := s.mq.PublishJSON(r.Context(), events.BookingCreatedQueue, msg); err != nil {
		_ = s.stock.Release(context.Background(), req.EventID)
		_ = s.store.CancelBooking(context.Background(), bookingID)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"booking_id": bookingID, "status": "PENDING"})
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("extra json")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func randID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
