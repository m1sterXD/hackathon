package service

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

const expectedFields = 16

type Service struct {
	path string
	data map[int64]University
}

type INNPair struct {
	INN        int64      `json:"inn"`
	University University `json:"university"`
}
type University struct {
	ExtendedRank int
	NameShort    string
	Name         string
	INN          string
	Country      string
	Profile      string

	InTraining bool

	ActualScore     float64
	ActualRank      int
	PredictedScore  float64
	ScoreNormalized float64

	Metrics Metrics
}

type Metrics struct {
	F1 float64
	F2 float64
	F3 float64
	F4 float64
	F5 float64
}

type MetricsRequest Metrics

func NewService(path string) *Service {
	return &Service{
		path: path,
	}
}

func (s *Service) Init() error {
	return s.loadData(s.path)
}

func (s *Service) loadData(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t' // если табы
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	s.data = make(map[int64]University)

	// 🔥 ПРОПУСК HEADER
	_, err = reader.Read()
	if err != nil {
		return err
	}
	_, err = reader.Read()
	if err != nil {
		return err
	}

	line := 0

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("read error:", err)
			continue
		}

		line++

		// 🔥 ЖЁСТКАЯ ПРОВЕРКА
		if len(record) != expectedFields {
			log.Printf("skip line %d: expected %d fields, got %d\n", line, expectedFields, len(record))
			continue
		}

		u, ok := parseLine(record)
		if !ok {
			log.Printf("skip line %d: parse error\n", line)
			continue
		}

		key := mustINN(record[3])
		if key == 0 {
			log.Printf("skip line %d: invalid INN\n", line)
			continue
		}

		s.data[key] = u
	}

	return nil
}
func parseLine(f []string) (University, bool) {
	return University{
		ExtendedRank: mustAtoi(f[0]),
		NameShort:    f[1],
		Name:         f[2],
		INN:          f[3],
		Country:      f[4],
		Profile:      f[5],

		InTraining: atob(f[6]),

		ActualScore:     mustAtof(f[7]),
		ActualRank:      mustAtoi(f[8]),
		PredictedScore:  mustAtof(f[9]),
		ScoreNormalized: mustAtof(f[10]),

		Metrics: Metrics{
			F1: mustAtof(f[11]),
			F2: mustAtof(f[12]),
			F3: mustAtof(f[13]),
			F4: mustAtof(f[14]),
			F5: mustAtof(f[15]),
		},
	}, true
}

func mustAtof(s string) float64 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Println("float parse error:", s)
		return 0
	}
	return v
}

func mustAtoi(s string) int {
	if s == "" {
		return 0
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Println("int parse error:", s)
		return 0
	}
	return int(f)
}

func mustINN(s string) int64 {
	if s == "" {
		return 0
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Println("INN parse error:", s)
		return 0
	}

	return int64(f)
}

func atob(s string) bool {
	return s == "1" || s == "true" || s == "TRUE"
}

func (s *Service) GetAllINNs() []INNPair {
	result := make([]INNPair, 0, len(s.data))

	for k, v := range s.data {
		result = append(result, INNPair{
			INN:        k,
			University: v,
		})
	}

	return result
}

func (s *Service) GetByINN(inn int64) (University, bool) {
	u, ok := s.data[inn]
	return u, ok
}
