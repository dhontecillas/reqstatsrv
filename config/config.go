package config

import (
	"fmt"
)

type Config struct {
	Port       int        `json:"port"`
	Host       string     `json:"host"`        //
	RandomSeed int64      `json:"random_seed"` // value to get reproducible results
	Endpoints  []Endpoint `json:"endpoints"`
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func DefaultConfig() *Config {
	return &Config{
		Port: 9876,
		Host: "0.0.0.0",
		Endpoints: []Endpoint{
			Endpoint{
				Method:      "GET",
				PathPattern: "/",
				Behaviour: []Behaviour{
					Behaviour{
						Name: "slower",
						Config: map[string]interface{}{
							"freq": 0.5,
							"seed": 1,
						},
					},
					Behaviour{
						Name: "status_distributor",
						Config: map[string]interface{}{
							"code_distribution": map[int]float64{
								200: 0.5,
								201: 0.3,
								202: 0.1,
								204: 0.1,
							},
							"seed": 1,
						},
					},
					Behaviour{
						Name: "connection_closer",
						Config: map[string]interface{}{
							"freq": 0.5,
							"seed": 1,
						},
					},
				},
				Content: Content{
					Source: "directory",
					Config: map[string]interface{}{
						"dir": "./example/data",
					},
				},
			},
		},
	}
}

// Endpoint defines an endpoint
type Endpoint struct {
	Method      string      `json:"method"`
	PathPattern string      `json:"path_pattern"`
	Behaviour   []Behaviour `json:"behaviour"`
	Content     Content     `json:"content"`
}

// Content defines what the body and headers
// will contain, by selecting the source for the
// payload, and providing a "free form" configuration
// to be interpreted by each kind of content source
type Content struct {
	Source string                 `json:"source"`
	Config map[string]interface{} `json:"config"`
}

// Behaviour defines a modification in how the response
// will behave in terms of errors, and status codes to produce
type Behaviour struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

type IntFloat struct {
	Key int     `json:"key"`
	Val float64 `json:"val"`
}

type IntDistribution []IntFloat
