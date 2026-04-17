package server

import (
	"encoding/json"
	"log"
	"main/service"
	"net/http"
	"strings"
)

type Server struct {
	svc *service.Service
}

func NewServer(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/universities", s.handleUniversities)
	mux.HandleFunc("/universities/", s.handleUniversityByName)
	mux.HandleFunc("/metrics", s.handleMetrics)

	return http.ListenAndServe(":8080", corsMiddleware(mux))
}

func (s *Server) handleUniversities(w http.ResponseWriter, r *http.Request) {
	log.Print("Handle Universities", r)
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data := s.svc.GetAllNames()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleUniversityByName(w http.ResponseWriter, r *http.Request) {
	log.Print("Handling university by name", r)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/universities/")

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uni, isExist := s.svc.GetByName(name)
	if !isExist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(uni)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req service.MetricsRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	avg := calculateAverage(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"average": avg,
	})
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
func calculateAverage(m service.MetricsRequest) float64 {
	values := []float64{
		m.F1.AcademicSelectivity,
		m.F1.OlympiadElite,
		m.F1.CompetitionPressure,

		m.F2.FinancialCapacity,
		m.F2.FacultyQuality,
		m.F2.TeachingIntensity,
		m.F2.ProgramDepth,

		m.F3.EmployerTrust,
		m.F3.IndustrySpread,
		m.F3.PremiumSegment,

		m.F4.Recognition,
		m.F4.SubjectStrength,
		m.F4.StaffIntl,
		m.F4.StudentIntl,
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}
