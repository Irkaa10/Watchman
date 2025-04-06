package config

import (
	"github.com/Irkaa10/Watchman/models"
)

func LoadConfig() models.Config {
	return models.Config{
		Port: "8080",
		Services: []models.Service{
			{
				Name:     "users-service",
				URL:      "http://localhost:8081",
				Prefixes: []string{"/users", "/auth"},
			},
			{
				Name:     "products-service",
				URL:      "http://localhost:8082",
				Prefixes: []string{"/products", "/categories"},
			},
		},
	}
}
