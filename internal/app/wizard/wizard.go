package wizard

import (
	"fmt"
	"os"
	"strings"
)

// RunWizard runs the interactive configuration wizard and returns the resulting StackConfig.
// existing may be nil for a first-time setup, or non-nil to pre-fill defaults for --reconfigure.
func RunWizard(p *Prompter, existing *StackConfig) (*StackConfig, error) {
	defaults := Defaults()
	if existing != nil {
		defaults = *existing
	}

	fmt.Fprintln(p.Out, "rulekit: first-time setup — let's configure your stack.")
	fmt.Fprintln(p.Out, "rulekit: press enter to accept defaults shown in [brackets].")
	fmt.Fprintln(p.Out)
	fmt.Fprintln(p.Out, "─────────────────────────────────────────")

	cfg := defaults

	// Step 1 — Database
	fmt.Fprintln(p.Out, "\nSTEP 1 — Database")
	fmt.Fprintln(p.Out, "\n  Which database would you like to use?")

	dbDefault := 0
	if defaults.Store == "postgres" {
		dbDefault = 1
	}
	dbIdx, err := p.PromptSelect("", []string{
		"SQLite   — zero config, great for local dev and small teams (default)",
		"Postgres — recommended for production and multi-user setups",
	}, dbDefault)
	if err != nil {
		return nil, err
	}

	if dbIdx == 0 {
		cfg.Store = "sqlite"
		cfg.DataDir = "/data"
		cfg.DatabaseURL = ""
	} else {
		cfg.Store = "postgres"
		pgDefault := defaults.DatabaseURL
		if pgDefault == "" {
			pgDefault = "postgres://rulekit:rulekit@localhost:5432/rulekit"
		}
		pgURL, err := promptValidated(p, "Postgres connection URL", pgDefault, validatePostgresURL)
		if err != nil {
			return nil, err
		}
		cfg.DatabaseURL = pgURL
	}

	fmt.Fprintln(p.Out, "\n─────────────────────────────────────────")

	// Step 2 — Blob storage
	fmt.Fprintln(p.Out, "\nSTEP 2 — Blob storage")
	fmt.Fprintln(p.Out, "\n  Where should rule bundles be stored?")

	blobDefault := 0
	if defaults.BlobStore == "s3" {
		blobDefault = 1
	}
	blobIdx, err := p.PromptSelect("", []string{
		"Filesystem — stored inside the container volume (default)",
		"S3         — AWS S3 or any S3-compatible store (Cloudflare R2, MinIO)",
	}, blobDefault)
	if err != nil {
		return nil, err
	}

	if blobIdx == 0 {
		cfg.BlobStore = "fs"
		cfg.S3Bucket = ""
		cfg.S3Region = ""
		cfg.S3Endpoint = ""
		cfg.S3AccessKeyID = ""
		cfg.S3SecretAccessKey = ""
	} else {
		cfg.BlobStore = "s3"

		bucket, err := promptValidated(p, "Bucket name", defaults.S3Bucket, func(s string) error {
			if s == "" {
				return fmt.Errorf("bucket name is required")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		cfg.S3Bucket = bucket

		region, err := p.PromptText("Region", orDefault(defaults.S3Region, "us-east-1"))
		if err != nil {
			return nil, err
		}
		cfg.S3Region = region

		endpoint, err := p.PromptText("Endpoint (leave blank for AWS)", defaults.S3Endpoint)
		if err != nil {
			return nil, err
		}
		cfg.S3Endpoint = endpoint

		accessKey, err := promptValidated(p, "Access key ID", defaults.S3AccessKeyID, func(s string) error {
			if s == "" {
				return fmt.Errorf("access key ID is required")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		cfg.S3AccessKeyID = accessKey

		secretKey, err := p.PromptSecret("Secret access key")
		if err != nil {
			return nil, err
		}
		if secretKey == "" && defaults.S3SecretAccessKey != "" {
			secretKey = defaults.S3SecretAccessKey
		}
		if secretKey == "" {
			return nil, fmt.Errorf("secret access key is required")
		}
		cfg.S3SecretAccessKey = secretKey
	}

	fmt.Fprintln(p.Out, "\n─────────────────────────────────────────")

	// Step 3 — Authentication
	fmt.Fprintln(p.Out, "\nSTEP 3 — Authentication")
	fmt.Fprintln(p.Out)

	adminPassword, err := p.PromptSecret("Admin password (leave blank to generate one)")
	if err != nil {
		return nil, err
	}
	if adminPassword == "" {
		if existing != nil && defaults.AdminPassword != "" {
			adminPassword = defaults.AdminPassword
			fmt.Fprintln(p.Out, "  rulekit: keeping existing admin password.")
		} else {
			generated, err := GenerateSecret(16)
			if err != nil {
				return nil, err
			}
			adminPassword = generated
			fmt.Fprintf(p.Out, "  rulekit: generated admin password: %s\n", adminPassword)
			fmt.Fprintln(p.Out, "  rulekit: save this — it will not be shown again.")
		}
	}
	cfg.AdminPassword = adminPassword

	if existing != nil && defaults.JWTSecret != "" {
		cfg.JWTSecret = defaults.JWTSecret
	} else {
		secret, err := GenerateSecret(32)
		if err != nil {
			return nil, err
		}
		cfg.JWTSecret = secret
	}

	configureSMTP, err := p.PromptConfirm("Configure SMTP? Without it, OTP codes print to registry stdout.", false)
	if err != nil {
		return nil, err
	}
	cfg.SMTPEnabled = configureSMTP
	if configureSMTP {
		if err := fillSMTP(p, &cfg, existing); err != nil {
			return nil, err
		}
	} else {
		cfg.SMTPHost = ""
		cfg.SMTPUsername = ""
		cfg.SMTPPassword = ""
	}

	fmt.Fprintln(p.Out, "\n─────────────────────────────────────────")

	// Step 4 — Ports
	fmt.Fprintln(p.Out, "\nSTEP 4 — Ports")
	fmt.Fprintln(p.Out)

	regPort, err := promptValidated(p,
		fmt.Sprintf("Registry port [%d]", defaults.RegistryPort),
		fmt.Sprintf("%d", defaults.RegistryPort),
		validatePort)
	if err != nil {
		return nil, err
	}
	cfg.RegistryPort = mustAtoi(regPort)

	dashPort, err := promptValidated(p,
		fmt.Sprintf("Dashboard port [%d]", defaults.DashboardPort),
		fmt.Sprintf("%d", defaults.DashboardPort),
		validatePort)
	if err != nil {
		return nil, err
	}
	cfg.DashboardPort = mustAtoi(dashPort)

	fmt.Fprintln(p.Out, "\n─────────────────────────────────────────")

	// Step 5 — Confirm
	fmt.Fprintln(p.Out, "\nSTEP 5 — Confirm")
	fmt.Fprintln(p.Out)
	fmt.Fprintln(p.Out, "rulekit: here's your configuration:")
	fmt.Fprintln(p.Out)
	fmt.Fprint(p.Out, cfg.Summary())
	fmt.Fprintln(p.Out)

	ok, err := p.PromptConfirm("Save and start?", true)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCancelled
	}

	return &cfg, nil
}

// RunWithDefaults builds a StackConfig from defaults without any prompts (--yes flag).
// JWTSecret and AdminPassword are always auto-generated.
func RunWithDefaults(p *Prompter) *StackConfig {
	cfg := Defaults()

	jwtSecret, _ := GenerateSecret(32)
	cfg.JWTSecret = jwtSecret

	adminPassword, _ := GenerateSecret(16)
	cfg.AdminPassword = adminPassword

	fmt.Fprintln(p.Out, "rulekit: using defaults (--yes). run 'rulekit onboard --reconfigure' to change.")
	fmt.Fprintf(p.Out, "rulekit: database       sqlite\n")
	fmt.Fprintf(p.Out, "rulekit: blob           filesystem\n")
	fmt.Fprintf(p.Out, "rulekit: admin password %s\n", adminPassword)
	fmt.Fprintf(p.Out, "rulekit: registry       http://localhost:%d\n", cfg.RegistryPort)
	fmt.Fprintf(p.Out, "rulekit: dashboard      http://localhost:%d\n", cfg.DashboardPort)
	return &cfg
}

// ErrCancelled is returned when the user aborts the wizard.
var ErrCancelled = fmt.Errorf("setup cancelled. nothing was saved")

// fillSMTP fills SMTP fields on cfg interactively.
func fillSMTP(p *Prompter, cfg *StackConfig, existing *StackConfig) error {
	smtpDefaults := &StackConfig{SMTPPort: 587, SMTPFrom: "noreply@rulekit.dev"}
	if existing != nil {
		smtpDefaults = existing
	}

	host, err := promptValidated(p, "SMTP host", smtpDefaults.SMTPHost, func(s string) error {
		if s == "" {
			return fmt.Errorf("SMTP host is required")
		}
		return nil
	})
	if err != nil {
		return err
	}
	cfg.SMTPHost = host

	port, err := promptValidated(p, "SMTP port", fmt.Sprintf("%d", smtpDefaults.SMTPPort), validateSMTPPort)
	if err != nil {
		return err
	}
	cfg.SMTPPort = mustAtoi(port)

	username, err := p.PromptText("SMTP username", smtpDefaults.SMTPUsername)
	if err != nil {
		return err
	}
	cfg.SMTPUsername = username

	password, err := p.PromptSecret("SMTP password")
	if err != nil {
		return err
	}
	if password == "" && existing != nil && smtpDefaults.SMTPPassword != "" {
		password = smtpDefaults.SMTPPassword
	}
	cfg.SMTPPassword = password

	fromAddr, err := p.PromptText("From address", orDefault(smtpDefaults.SMTPFrom, "noreply@rulekit.dev"))
	if err != nil {
		return err
	}
	cfg.SMTPFrom = fromAddr

	useTLS, err := p.PromptConfirm("Use TLS?", false)
	if err != nil {
		return err
	}
	cfg.SMTPUseTLS = useTLS

	return nil
}

// promptValidated re-prompts until the validation function returns nil.
func promptValidated(p *Prompter, question, defaultVal string, validate func(string) error) (string, error) {
	for {
		val, err := p.PromptText(question, defaultVal)
		if err != nil {
			return "", err
		}
		if err := validate(val); err != nil {
			fmt.Fprintf(p.Out, "  rulekit: error: %v\n", err)
			if !IsTTY() {
				return "", err
			}
			continue
		}
		return val, nil
	}
}

// Validation helpers

func validatePostgresURL(s string) error {
	if !strings.HasPrefix(s, "postgres://") && !strings.HasPrefix(s, "postgresql://") {
		return fmt.Errorf("must start with postgres:// or postgresql://")
	}
	return nil
}

func validatePort(s string) error {
	n := mustAtoi(s)
	if n < 1024 || n > 65535 {
		return fmt.Errorf("must be between 1024 and 65535")
	}
	return nil
}

func validateSMTPPort(s string) error {
	n := mustAtoi(s)
	if n < 1 || n > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}
	return nil
}

func mustAtoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func orDefault(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}

// EnvExists returns true if the .env file already exists at path.
func EnvExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
