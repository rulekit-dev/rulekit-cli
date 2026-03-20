package docker

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ComposeOptions configures the generated docker-compose.yml.
type ComposeOptions struct {
	RegistryPort    int
	DashboardPort   int
	UsePostgres     bool
	RegistryImage   string
	DashboardImage  string
}

// DefaultComposeOptions returns the default options matching the spec.
func DefaultComposeOptions() ComposeOptions {
	return ComposeOptions{
		RegistryPort:   8080,
		DashboardPort:  3000,
		RegistryImage:  "ghcr.io/rulekit-dev/rulekit-registry:latest",
		DashboardImage: "ghcr.io/rulekit-dev/rulekit-dashboard:latest",
	}
}

// composeFile is the top-level docker-compose.yml structure.
type composeFile struct {
	Version  string                 `yaml:"version"`
	Services map[string]service     `yaml:"services"`
	Volumes  map[string]emptyVolume `yaml:"volumes,omitempty"`
}

// emptyVolume represents a named Docker volume declaration (no config needed).
type emptyVolume struct{}

type service struct {
	Image       string               `yaml:"image"`
	EnvFile     []string             `yaml:"env_file,omitempty"`
	Ports       []string             `yaml:"ports,omitempty"`
	Environment map[string]string    `yaml:"environment,omitempty"`
	Volumes     []string             `yaml:"volumes,omitempty"`
	HealthCheck *healthCheck         `yaml:"healthcheck,omitempty"`
	DependsOn   map[string]condition `yaml:"depends_on,omitempty"`
}

type healthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

type condition struct {
	Condition string `yaml:"condition"`
}

// GenerateCompose builds a ComposeFile struct and writes it atomically to ComposePath().
func GenerateCompose(opts ComposeOptions) error {
	cf := buildCompose(opts)

	data, err := yaml.Marshal(cf)
	if err != nil {
		return fmt.Errorf("marshal compose: %w", err)
	}

	dir := ComposeDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create compose dir: %w", err)
	}

	dest := ComposePath()
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write compose tmp: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("rename compose: %w", err)
	}

	return nil
}

func buildCompose(opts ComposeOptions) composeFile {
	registryDependsOn := map[string]condition{}
	volumes := map[string]emptyVolume{
		"registry-data": {},
	}

	if opts.UsePostgres {
		registryDependsOn["postgres"] = condition{Condition: "service_healthy"}
		volumes["postgres-data"] = emptyVolume{}
	}

	// All registry configuration flows through the .env file (env_file: .env).
	registrySvc := service{
		Image:   opts.RegistryImage,
		EnvFile: []string{".env"},
		Ports:   []string{fmt.Sprintf("%d:8080", opts.RegistryPort)},
		HealthCheck: &healthCheck{
			Test:     []string{"CMD", "wget", "-qO-", "http://localhost:8080/healthz"},
			Interval: "5s",
			Timeout:  "3s",
			Retries:  5,
		},
	}

	if opts.UsePostgres {
		registrySvc.DependsOn = registryDependsOn
	} else {
		registrySvc.Volumes = []string{"registry-data:/data"}
	}

	dashboardSvc := service{
		Image: opts.DashboardImage,
		Ports: []string{fmt.Sprintf("%d:3000", opts.DashboardPort)},
		Environment: map[string]string{
			"RULEKIT_REGISTRY_URL": "http://registry:8080",
		},
		DependsOn: map[string]condition{
			"registry": {Condition: "service_healthy"},
		},
	}

	services := map[string]service{
		"registry":  registrySvc,
		"dashboard": dashboardSvc,
	}

	if opts.UsePostgres {
		services["postgres"] = service{
			Image: "postgres:16-alpine",
			Environment: map[string]string{
				"POSTGRES_DB":       "rulekit",
				"POSTGRES_USER":     "rulekit",
				"POSTGRES_PASSWORD": "rulekit",
			},
			Volumes: []string{"postgres-data:/var/lib/postgresql/data"},
			HealthCheck: &healthCheck{
				Test:     []string{"CMD", "pg_isready", "-U", "rulekit"},
				Interval: "5s",
				Timeout:  "3s",
				Retries:  5,
			},
		}
	}

	return composeFile{
		Version:  "3.9",
		Services: services,
		Volumes:  volumes,
	}
}

// ParseDatabaseType reads the compose file at composePath and returns "postgres" or "sqlite".
func ParseDatabaseType(composePath string) string {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return "sqlite"
	}
	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return "sqlite"
	}
	if _, ok := cf.Services["postgres"]; ok {
		return "postgres"
	}
	return "sqlite"
}
