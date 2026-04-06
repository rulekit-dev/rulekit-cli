package wizard

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ErrCancelled is returned when the user aborts the wizard.
var ErrCancelled = fmt.Errorf("setup cancelled. nothing was saved")

// RunWizard runs the interactive configuration wizard and returns the resulting StackConfig.
// existing may be nil for a first-time setup, or non-nil to pre-fill defaults for --reconfigure.
func RunWizard(p *Prompter, existing *StackConfig) (*StackConfig, error) {
	if !IsTTY() {
		return runWizardFallback(p, existing)
	}
	return runWizardInteractive(existing)
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

// runWizardInteractive runs the huh-powered interactive wizard.
func runWizardInteractive(existing *StackConfig) (*StackConfig, error) {
	defaults := Defaults()
	if existing != nil {
		defaults = *existing
	}
	cfg := defaults

	theme := orangeTheme()

	// ── Step 1: Database ──────────────────────────────────────────────────────

	dbStore := cfg.Store
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = "postgres://rulekit:rulekit@localhost:5432/rulekit"
	}

	dbGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Database").
			Description("Where should rulekit store its data?").
			Options(
				huh.NewOption("SQLite   — zero config, great for local dev", "sqlite"),
				huh.NewOption("Postgres — recommended for production", "postgres"),
			).
			Value(&dbStore),
	)

	if err := huh.NewForm(dbGroup).WithTheme(theme).Run(); err != nil {
		return nil, errOrCancelled(err)
	}

	cfg.Store = dbStore

	if dbStore == "postgres" {
		pgGroup := huh.NewGroup(
			huh.NewInput().
				Title("Postgres connection URL").
				Value(&dbURL).
				Validate(func(s string) error {
					if !strings.HasPrefix(s, "postgres://") && !strings.HasPrefix(s, "postgresql://") {
						return fmt.Errorf("must start with postgres:// or postgresql://")
					}
					return nil
				}),
		)
		if err := huh.NewForm(pgGroup).WithTheme(theme).Run(); err != nil {
			return nil, errOrCancelled(err)
		}
		cfg.DatabaseURL = dbURL
		cfg.DataDir = ""
	} else {
		cfg.DataDir = "/data"
		cfg.DatabaseURL = ""
	}

	// ── Step 2: Blob storage ──────────────────────────────────────────────────

	blobStore := cfg.BlobStore

	blobGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Blob storage").
			Description("Where should rule bundles be stored?").
			Options(
				huh.NewOption("Filesystem — stored in the container volume", "fs"),
				huh.NewOption("S3         — AWS S3 or S3-compatible (R2, MinIO)", "s3"),
			).
			Value(&blobStore),
	)
	if err := huh.NewForm(blobGroup).WithTheme(theme).Run(); err != nil {
		return nil, errOrCancelled(err)
	}

	cfg.BlobStore = blobStore

	if blobStore == "s3" {
		s3Bucket := cfg.S3Bucket
		s3Region := orDefault(cfg.S3Region, "us-east-1")
		s3Endpoint := cfg.S3Endpoint
		s3AccessKey := cfg.S3AccessKeyID
		s3SecretKey := cfg.S3SecretAccessKey

		s3Group := huh.NewGroup(
			huh.NewInput().
				Title("Bucket name").
				Value(&s3Bucket).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("bucket name is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Region").
				Value(&s3Region),
			huh.NewInput().
				Title("Endpoint").
				Description("Leave blank for AWS S3").
				Value(&s3Endpoint),
			huh.NewInput().
				Title("Access key ID").
				Value(&s3AccessKey).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("access key ID is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Secret access key").
				EchoMode(huh.EchoModePassword).
				Value(&s3SecretKey).
				Validate(func(s string) error {
					if s == "" && cfg.S3SecretAccessKey == "" {
						return fmt.Errorf("secret access key is required")
					}
					return nil
				}),
		)
		if err := huh.NewForm(s3Group).WithTheme(theme).Run(); err != nil {
			return nil, errOrCancelled(err)
		}

		cfg.S3Bucket = s3Bucket
		cfg.S3Region = s3Region
		cfg.S3Endpoint = s3Endpoint
		cfg.S3AccessKeyID = s3AccessKey
		if s3SecretKey != "" {
			cfg.S3SecretAccessKey = s3SecretKey
		}
	} else {
		cfg.S3Bucket = ""
		cfg.S3Region = ""
		cfg.S3Endpoint = ""
		cfg.S3AccessKeyID = ""
		cfg.S3SecretAccessKey = ""
	}

	// ── Step 3: Auth ──────────────────────────────────────────────────────────

	adminPassword := ""
	keepPassword := existing != nil && defaults.AdminPassword != ""

	passwordNote := "Leave blank to auto-generate a secure password."
	if keepPassword {
		passwordNote = "Leave blank to keep your existing password."
	}

	authGroup := huh.NewGroup(
		huh.NewInput().
			Title("Admin password").
			Description(passwordNote).
			EchoMode(huh.EchoModePassword).
			Value(&adminPassword),
	)
	if err := huh.NewForm(authGroup).WithTheme(theme).Run(); err != nil {
		return nil, errOrCancelled(err)
	}

	if adminPassword == "" {
		if keepPassword {
			cfg.AdminPassword = defaults.AdminPassword
		} else {
			generated, err := GenerateSecret(16)
			if err != nil {
				return nil, err
			}
			cfg.AdminPassword = generated
		}
	} else {
		cfg.AdminPassword = adminPassword
	}

	if existing != nil && defaults.JWTSecret != "" {
		cfg.JWTSecret = defaults.JWTSecret
	} else {
		secret, err := GenerateSecret(32)
		if err != nil {
			return nil, err
		}
		cfg.JWTSecret = secret
	}

	// ── Step 4: SMTP ──────────────────────────────────────────────────────────

	configureSMTP := cfg.SMTPEnabled

	smtpGroup := huh.NewGroup(
		huh.NewConfirm().
			Title("Configure SMTP?").
			Description("Without it, OTP codes print to registry stdout.").
			Value(&configureSMTP),
	)
	if err := huh.NewForm(smtpGroup).WithTheme(theme).Run(); err != nil {
		return nil, errOrCancelled(err)
	}

	cfg.SMTPEnabled = configureSMTP

	if configureSMTP {
		smtpHost := cfg.SMTPHost
		smtpPort := fmt.Sprintf("%d", orDefaultInt(cfg.SMTPPort, 587))
		smtpUser := cfg.SMTPUsername
		smtpPass := cfg.SMTPPassword
		smtpFrom := orDefault(cfg.SMTPFrom, "noreply@rulekit.dev")
		smtpTLS := cfg.SMTPUseTLS

		smtpDetails := huh.NewGroup(
			huh.NewInput().
				Title("SMTP host").
				Value(&smtpHost).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("SMTP host is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("SMTP port").
				Value(&smtpPort).
				Validate(validatePortStr),
			huh.NewInput().
				Title("SMTP username").
				Value(&smtpUser),
			huh.NewInput().
				Title("SMTP password").
				EchoMode(huh.EchoModePassword).
				Value(&smtpPass),
			huh.NewInput().
				Title("From address").
				Value(&smtpFrom),
			huh.NewConfirm().
				Title("Use TLS?").
				Value(&smtpTLS),
		)
		if err := huh.NewForm(smtpDetails).WithTheme(theme).Run(); err != nil {
			return nil, errOrCancelled(err)
		}

		cfg.SMTPHost = smtpHost
		cfg.SMTPPort = mustAtoi(smtpPort)
		cfg.SMTPUsername = smtpUser
		if smtpPass != "" {
			cfg.SMTPPassword = smtpPass
		}
		cfg.SMTPFrom = smtpFrom
		cfg.SMTPUseTLS = smtpTLS
	} else {
		cfg.SMTPHost = ""
		cfg.SMTPUsername = ""
		cfg.SMTPPassword = ""
	}

	// ── Step 5: Ports ─────────────────────────────────────────────────────────

	regPort := fmt.Sprintf("%d", cfg.RegistryPort)
	dashPort := fmt.Sprintf("%d", cfg.DashboardPort)

	portsGroup := huh.NewGroup(
		huh.NewInput().
			Title("Registry port").
			Value(&regPort).
			Validate(validatePortStr),
		huh.NewInput().
			Title("Dashboard port").
			Value(&dashPort).
			Validate(validatePortStr),
	)
	if err := huh.NewForm(portsGroup).WithTheme(theme).Run(); err != nil {
		return nil, errOrCancelled(err)
	}

	cfg.RegistryPort = mustAtoi(regPort)
	cfg.DashboardPort = mustAtoi(dashPort)

	// ── Step 6: Confirm ───────────────────────────────────────────────────────

	printSummary(&cfg)

	confirm := true
	confirmGroup := huh.NewGroup(
		huh.NewConfirm().
			Title("Save configuration?").
			Value(&confirm),
	)
	if err := huh.NewForm(confirmGroup).WithTheme(theme).Run(); err != nil {
		return nil, errOrCancelled(err)
	}

	if !confirm {
		return nil, ErrCancelled
	}

	if adminPassword == "" && !keepPassword {
		fmt.Fprintf(os.Stdout, "\n%s admin password: %s\n%s\n\n",
			lipgloss.NewStyle().Foreground(orange).Bold(true).Render("✓"),
			lipgloss.NewStyle().Bold(true).Render(cfg.AdminPassword),
			lipgloss.NewStyle().Faint(true).Render("  Save this — it will not be shown again."),
		)
	}

	return &cfg, nil
}

// runWizardFallback is the non-TTY path (CI, piped input) — returns defaults.
func runWizardFallback(p *Prompter, existing *StackConfig) (*StackConfig, error) {
	cfg := Defaults()
	if existing != nil {
		cfg = *existing
	}

	if cfg.JWTSecret == "" {
		secret, err := GenerateSecret(32)
		if err != nil {
			return nil, err
		}
		cfg.JWTSecret = secret
	}
	if cfg.AdminPassword == "" {
		pass, err := GenerateSecret(16)
		if err != nil {
			return nil, err
		}
		cfg.AdminPassword = pass
	}

	fmt.Fprintln(p.Out, "rulekit: non-interactive mode — using existing or default config.")
	return &cfg, nil
}

// printSummary renders a styled summary box before the confirm prompt.
func printSummary(cfg *StackConfig) {
	title := lipgloss.NewStyle().Foreground(orange).Bold(true)
	label := lipgloss.NewStyle().Faint(true).Width(14)
	value := lipgloss.NewStyle()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(orange).
		Padding(0, 2).
		MarginTop(1).
		MarginBottom(1)

	var lines []string
	lines = append(lines, title.Render("Configuration summary"))
	lines = append(lines, "")

	dbVal := "sqlite"
	if cfg.Store == "postgres" {
		dbVal = "postgres · " + maskURL(cfg.DatabaseURL)
	}
	lines = append(lines, label.Render("database")+value.Render(dbVal))

	blobVal := "filesystem"
	if cfg.BlobStore == "s3" {
		blobVal = fmt.Sprintf("s3 · %s (%s)", cfg.S3Bucket, cfg.S3Region)
	}
	lines = append(lines, label.Render("blob")+value.Render(blobVal))

	if cfg.SMTPEnabled {
		lines = append(lines, label.Render("smtp")+value.Render(fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)))
	}

	lines = append(lines, label.Render("registry")+value.Render(fmt.Sprintf("http://localhost:%d", cfg.RegistryPort)))
	lines = append(lines, label.Render("dashboard")+value.Render(fmt.Sprintf("http://localhost:%d", cfg.DashboardPort)))

	fmt.Println(box.Render(strings.Join(lines, "\n")))
}

// orangeTheme returns a huh theme using rulekit's orange brand color.
func orangeTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Title = t.Focused.Title.Foreground(orange).Bold(true)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(orange)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(orange)
	t.Focused.Base = t.Focused.Base.BorderForeground(orange)
	t.Focused.Description = t.Focused.Description.Faint(true)
	return t
}

var orange = lipgloss.Color("#FF7800")

func errOrCancelled(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrCancelled
	}
	return err
}

// fillSMTP is no longer used (huh handles it inline), kept for compatibility.

// Validation helpers

func validatePortStr(s string) error {
	n := mustAtoi(s)
	if n < 1024 || n > 65535 {
		return fmt.Errorf("must be between 1024 and 65535")
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

func orDefaultInt(val, fallback int) int {
	if val != 0 {
		return val
	}
	return fallback
}

// EnvExists returns true if the .env file already exists at path.
func EnvExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
