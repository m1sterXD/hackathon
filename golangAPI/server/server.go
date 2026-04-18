package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"main/service"
	"net/http"
)

type Server struct {
	svc *service.Service
}

type MetricsRequest struct {
	F1 float64 `json:"f1"`
	F2 float64 `json:"f2"`
	F3 float64 `json:"f3"`
	F4 float64 `json:"f4"`
	F5 float64 `json:"f5"`
}

type SearchRequest struct {
	INN int64 `json:"inn"`
}

func NewServer(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/universities", s.handleUniversities)
	mux.HandleFunc("/post", s.handleUniversitySearch)
	mux.HandleFunc("/forteen", s.handle14feats)

	return http.ListenAndServe(":8080", corsMiddleware(mux))
}

func (s *Server) handleUniversities(w http.ResponseWriter, r *http.Request) {
	log.Print("Handle Universities", r)
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data := s.svc.GetAllINNs()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) handle14feats(w http.ResponseWriter, r *http.Request) {
	log.Print("handle14feats")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// пробрасываем в python
	resp, err := http.Post(
		"http://pythonapi:8000/data_ars_forteen",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Println("python request error:", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("python read error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (s *Server) handleUniversitySearch(w http.ResponseWriter, r *http.Request) {
	log.Print("Handling university search")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uni, ok := s.svc.GetByINN(req.INN)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 🔥 ВОТ И ВСЁ — возвращаем полностью
	json.NewEncoder(w).Encode(uni)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
