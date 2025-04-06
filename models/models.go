package models

type Service struct {
	Name     string
	URL      string
	Prefixes []string
}

type Config struct {
	Port     string
	Services []Service
}
