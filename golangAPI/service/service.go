package service

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

type Service struct {
	path string
	data map[string]University
}

type University struct {
	Id   string
	Name string

	Lat float64
	Lon float64

	Metrics Metrics
}

type Metrics struct {
	F1 HumanCapital
	F2 Institution
	F3 Market
	F4 Global
}

type HumanCapital struct {
	AcademicSelectivity float64
	OlympiadElite       float64
	CompetitionPressure float64
}

type Institution struct {
	FinancialCapacity float64
	FacultyQuality    float64
	TeachingIntensity float64
	ProgramDepth      float64
}

type Market struct {
	EmployerTrust  float64
	IndustrySpread float64
	PremiumSegment float64
}

type Global struct {
	Recognition     float64
	SubjectStrength float64
	StaffIntl       float64
	StudentIntl     float64
}

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

	s.data = make(map[string]University)

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		u := parseLine(record)
		s.data[u.Name] = u
	}

	return nil
}

func parseLine(f []string) University {
	return University{
		Id:   f[0],
		Name: f[1],

		Lat: atof(f[2]),
		Lon: atof(f[3]),

		Metrics: Metrics{
			F1: HumanCapital{
				AcademicSelectivity: atof(f[4]),
				OlympiadElite:       atof(f[5]),
				CompetitionPressure: atof(f[6]),
			},
			F2: Institution{
				FinancialCapacity: atof(f[7]),
				FacultyQuality:    atof(f[8]),
				TeachingIntensity: atof(f[9]),
				ProgramDepth:      atof(f[10]),
			},
			F3: Market{
				EmployerTrust:  atof(f[11]),
				IndustrySpread: atof(f[12]),
				PremiumSegment: atof(f[13]),
			},
			F4: Global{
				Recognition:     atof(f[14]),
				SubjectStrength: atof(f[15]),
				StaffIntl:       atof(f[16]),
				StudentIntl:     atof(f[17]),
			},
		},
	}
}

func atof(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func (s *Service) GetAllNames() []string {
	names := make([]string, 0, len(s.data))
	for _, d := range s.data {
		names = append(names, d.Name)
	}
	return names
}

func (s *Service) GetByName(name string) (University, bool) {
	u, ok := s.data[name]
	return u, ok
}
