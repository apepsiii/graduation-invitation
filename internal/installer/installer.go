package installer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	resources "undangan-digital"
)

type Config struct {
	AppName      string `yaml:"app_name"`
	Port         int    `yaml:"port"`
	Instance     int    `yaml:"instance"`
	InstallDir   string `yaml:"install_dir"`
	AdminUser    string `yaml:"admin_user"`
	AdminPass    string `yaml:"admin_pass"`
	SessionSecret string `yaml:"session_secret"`
	CreatedAt    string `yaml:"created_at"`
	Version      string `yaml:"version"`
	BinaryPath   string `yaml:"binary_path"`
}

type Installer struct {
	config     *Config
	configPath string
	binaryPath string
}

func NewInstaller() *Installer {
	return &Installer{
		configPath: "/etc/undangan-digital.yaml",
		binaryPath: "/usr/local/bin/undangan-digital",
	}
}

func (i *Installer) Run() {
	if runtime.GOOS != "linux" {
		fmt.Println("⚠ Installer hanya berjalan di Linux (VPS)")
		fmt.Println("Untuk development lokal, gunakan: go run cmd/main.go serve")
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "serve" {
		i.runServer()
		return
	}

	configExists := i.loadConfig()
	if !configExists {
		i.showInstallWizard()
	} else {
		i.showMainMenu()
	}
}

func (i *Installer) loadConfig() bool {
	data, err := os.ReadFile(i.configPath)
	if err != nil {
		return false
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return false
	}
	i.config = &cfg
	return true
}

func (i *Installer) saveConfig() error {
	data, err := yaml.Marshal(i.config)
	if err != nil {
		return err
	}
	return os.WriteFile(i.configPath, data, 0644)
}

func (i *Installer) showBanner() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║    Undangan Digital - Installer v1.0    ║")
	fmt.Println("║    Platform Undangan Digital untuk      ║")
	fmt.Println("║    Sekolah & Acara                      ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
}

func (i *Installer) showMainMenu() {
	i.showBanner()
	fmt.Printf("Aplikasi: %s (Port: %d, Instance: %d)\n", i.config.AppName, i.config.Port, i.config.Instance)
	fmt.Printf("Status: %s\n", i.getServiceStatus())
	fmt.Println()
	fmt.Println("Menu:")
	fmt.Println("  1. Install Baru")
	fmt.Println("  2. Update Aplikasi")
	fmt.Println("  3. Lihat Status Service")
	fmt.Println("  4. Restart Service")
	fmt.Println("  5. Stop Service")
	fmt.Println("  6. Uninstall")
	fmt.Println("  7. Keluar")
	fmt.Println()

	choice := i.prompt("Pilihan [1-7]: ")
	switch choice {
	case "1":
		i.showInstallWizard()
	case "2":
		i.updateApp()
	case "3":
		i.showStatus()
	case "4":
		i.restartService()
	case "5":
		i.stopService()
	case "6":
		i.uninstall()
	case "7":
		fmt.Println("Terima kasih!")
	default:
		fmt.Println("Pilihan tidak valid")
	}
}

func (i *Installer) showInstallWizard() {
	i.showBanner()
	fmt.Println("=== INSTALL BARU ===")
	fmt.Println()

	appName := i.prompt("Nama aplikasi [undangan-digital]: ")
	if appName == "" {
		appName = "undangan-digital"
	}

	portStr := i.prompt("Port [8080]: ")
	port := 8080
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err == nil && p > 0 && p < 65535 {
			port = p
		}
	}

	instanceStr := i.prompt("Instance keberapa [1]: ")
	instance := 1
	if instanceStr != "" {
		in, err := strconv.Atoi(instanceStr)
		if err == nil && in > 0 {
			instance = in
		}
	}

	adminUser := i.prompt("Username admin [admin]: ")
	if adminUser == "" {
		adminUser = "admin"
	}

	adminPass := i.prompt("Password admin [admin123]: ")
	if adminPass == "" {
		adminPass = "admin123"
	}

	installDir := fmt.Sprintf("/opt/%s-%d", appName, instance)
	if instance == 1 {
		installDir = fmt.Sprintf("/opt/%s", appName)
	}

	fmt.Println()
	fmt.Println("=== Konfirmasi ===")
	fmt.Printf("Nama aplikasi: %s\n", appName)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Instance: %d\n", instance)
	fmt.Printf("Install dir: %s\n", installDir)
	fmt.Printf("Admin: %s / %s\n", adminUser, adminPass)
	fmt.Println()

	confirm := i.prompt("Lanjutkan install? [y/N]: ")
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Install dibatalkan")
		return
	}

	i.config = &Config{
		AppName:      appName,
		Port:         port,
		Instance:     instance,
		InstallDir:   installDir,
		AdminUser:    adminUser,
		AdminPass:    adminPass,
		SessionSecret: generateSecret(),
		CreatedAt:    time.Now().Format(time.RFC3339),
		Version:      resources.Version,
		BinaryPath:   i.binaryPath,
	}

	i.install()
}

func (i *Installer) install() {
	fmt.Println()
	fmt.Println("=== Proses Install ===")

	fmt.Print("1. Membuat direktori install... ")
	if err := os.MkdirAll(i.config.InstallDir, 0755); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	if err := os.MkdirAll(filepath.Join(i.config.InstallDir, "database"), 0755); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	if err := os.MkdirAll(filepath.Join(i.config.InstallDir, "assets", "uploads"), 0755); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("2. Extract templates... ")
	if err := extractFS(resources.Templates(), "", filepath.Join(i.config.InstallDir, "templates")); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("3. Extract assets... ")
	if err := extractFS(resources.Assets(), "", filepath.Join(i.config.InstallDir, "assets")); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("4. Copy binary... ")
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	if err := copyFile(exePath, i.binaryPath); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("5. Save config... ")
	if err := i.saveConfig(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("6. Create systemd service... ")
	if err := i.createSystemdService(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("7. Enable service... ")
	if err := exec.Command("systemctl", "enable", i.config.AppName).Run(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("8. Start service... ")
	if err := exec.Command("systemctl", "start", i.config.AppName).Run(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║          INSTALL BERHASIL!              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Aplikasi: %s\n", i.config.AppName)
	fmt.Printf("URL: http://<IP-VPS>:%d\n", i.config.Port)
	fmt.Printf("Admin: http://<IP-VPS>:%d/admin/login\n", i.config.Port)
	fmt.Printf("Login: %s / %s\n", i.config.AdminUser, i.config.AdminPass)
	fmt.Printf("Install dir: %s\n", i.config.InstallDir)
	fmt.Printf("Config: %s\n", i.configPath)
	fmt.Println()
}

func (i *Installer) updateApp() {
	fmt.Println()
	fmt.Println("=== UPDATE APLIKASI ===")

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	fmt.Print("1. Stop service... ")
	exec.Command("systemctl", "stop", i.config.AppName).Run()
	fmt.Println("OK")

	fmt.Print("2. Update binary... ")
	if err := copyFile(exePath, i.binaryPath); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("3. Update templates... ")
	if err := extractFS(resources.Templates(), "", filepath.Join(i.config.InstallDir, "templates")); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("4. Update assets... ")
	if err := extractFS(resources.Assets(), "", filepath.Join(i.config.InstallDir, "assets")); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Print("5. Update config version... ")
	i.config.Version = resources.Version
	i.saveConfig()
	fmt.Println("OK")

	fmt.Print("6. Start service... ")
	if err := exec.Command("systemctl", "start", i.config.AppName).Run(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")

	fmt.Println()
	fmt.Println("✓ Update berhasil! Service sudah berjalan.")
}

func (i *Installer) showStatus() {
	fmt.Println()
	fmt.Println("=== STATUS SERVICE ===")
	fmt.Printf("Nama: %s\n", i.config.AppName)
	fmt.Printf("Port: %d\n", i.config.Port)
	fmt.Printf("Status: %s\n", i.getServiceStatus())
	fmt.Printf("Install dir: %s\n", i.config.InstallDir)
	fmt.Printf("Version: %s\n", i.config.Version)
	fmt.Printf("Created: %s\n", i.config.CreatedAt)
	fmt.Println()

	out, _ := exec.Command("systemctl", "status", i.config.AppName, "--no-pager").Output()
	fmt.Println(string(out))
}

func (i *Installer) restartService() {
	fmt.Print("Restarting service... ")
	if err := exec.Command("systemctl", "restart", i.config.AppName).Run(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")
	fmt.Printf("Service %s berhasil di-restart\n", i.config.AppName)
}

func (i *Installer) stopService() {
	fmt.Print("Stopping service... ")
	if err := exec.Command("systemctl", "stop", i.config.AppName).Run(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("OK")
	fmt.Printf("Service %s berhasil di-stop\n", i.config.AppName)
}

func (i *Installer) uninstall() {
	fmt.Println()
	fmt.Println("=== UNINSTALL ===")
	confirm := i.prompt("Yakin ingin uninstall? Data akan dihapus! [y/N]: ")
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Uninstall dibatalkan")
		return
	}

	fmt.Print("1. Stop service... ")
	exec.Command("systemctl", "stop", i.config.AppName).Run()
	fmt.Println("OK")

	fmt.Print("2. Disable service... ")
	exec.Command("systemctl", "disable", i.config.AppName).Run()
	fmt.Println("OK")

	fmt.Print("3. Remove systemd file... ")
	os.Remove(fmt.Sprintf("/etc/systemd/system/%s.service", i.config.AppName))
	exec.Command("systemctl", "daemon-reload").Run()
	fmt.Println("OK")

	fmt.Print("4. Remove binary... ")
	os.Remove(i.binaryPath)
	fmt.Println("OK")

	fmt.Print("5. Remove install dir... ")
	os.RemoveAll(i.config.InstallDir)
	fmt.Println("OK")

	fmt.Print("6. Remove config... ")
	os.Remove(i.configPath)
	fmt.Println("OK")

	fmt.Println()
	fmt.Println("✓ Uninstall berhasil!")
}

func (i *Installer) getServiceStatus() string {
	out, err := exec.Command("systemctl", "is-active", i.config.AppName).Output()
	if err != nil {
		return "NOT INSTALLED"
	}
	return strings.TrimSpace(string(out))
}

func (i *Installer) createSystemdService() error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=%s - Undangan Digital
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s serve
Restart=always
RestartSec=5
Environment=DATABASE_PATH=%s/database/database.sqlite
Environment=PORT=%d
Environment=ADMIN_USER=%s
Environment=ADMIN_PASS=%s
Environment=SESSION_SECRET=%s

[Install]
WantedBy=multi-user.target
`, i.config.AppName, i.config.InstallDir, i.binaryPath, i.config.InstallDir, i.config.Port, i.config.AdminUser, i.config.AdminPass, i.config.SessionSecret)

	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", i.config.AppName)
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return err
	}
	return exec.Command("systemctl", "daemon-reload").Run()
}

func (i *Installer) prompt(msg string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(msg)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (i *Installer) runServer() {
	if i.config == nil {
		fmt.Println("ERROR: Config tidak ditemukan. Jalankan installer terlebih dahulu.")
		os.Exit(1)
	}

	os.Setenv("DATABASE_PATH", filepath.Join(i.config.InstallDir, "database", "database.sqlite"))
	os.Setenv("PORT", strconv.Itoa(i.config.Port))
	os.Setenv("ADMIN_USER", i.config.AdminUser)
	os.Setenv("ADMIN_PASS", i.config.AdminPass)
	os.Setenv("SESSION_SECRET", i.config.SessionSecret)

	os.Chdir(i.config.InstallDir)

	fmt.Printf("Server starting on port %d\n", i.config.Port)
	fmt.Printf("Admin: http://localhost:%d/admin/login\n", i.config.Port)
	fmt.Printf("Login: %s / %s\n", i.config.AdminUser, i.config.AdminPass)
}