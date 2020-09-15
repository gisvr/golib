package prom

// PrometheusConfig for Prometheus
type PrometheusOption struct {
	Enable bool   `yaml:"enable"`
	Addr   string `yaml:"addr"`
}
