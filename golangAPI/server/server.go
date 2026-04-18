package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/service"
	"net/http"
	"os"
	"sort"
)

var GROQ_KEY = os.Getenv("GROQ_KEY")

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

type AnalyzeRequest struct {
	Name    string    `json:"name"`
	Score   float64   `json:"score"`
	Metrics []float64 `json:"metrics"`
}

func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/universities", s.handleUniversities)
	mux.HandleFunc("/post", s.handleUniversitySearch)
	mux.HandleFunc("/forteen", s.handle14feats)
	mux.HandleFunc("/analyze", s.handleAnalyze)

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

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	log.Println("handleAnalyze")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// сортировка слабых метрик
	type pair struct {
		i int
		v float64
	}

	var sorted []pair
	for i, v := range req.Metrics {
		sorted = append(sorted, pair{i, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].v < sorted[j].v
	})

	weak := ""
	for i := 0; i < 2 && i < len(sorted); i++ {
		weak += fmt.Sprintf("F%d=%.1f ", sorted[i].i+1, sorted[i].v)
	}

	// формируем prompt
	prompt := fmt.Sprintf(
		`Ты аналитик рейтингов университетов.
ВУЗ: "%s".
Балл %.1f/100.
Слабейшие: %s.
Верни строго JSON: {"measures":[{"title":"до 5 слов","focus":"F1","action":"...","impact":"..."}]}`,
		req.Name,
		req.Score,
		weak,
	)

	payload := map[string]any{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]string{
			{"role": "system", "content": "Отвечай только JSON без текста."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
		"max_tokens":  800,
	}

	buf, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest(
		"POST",
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewBuffer(buf),
	)

	httpReq.Header.Set("Authorization", "Bearer "+GROQ_KEY)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
