package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strings"
)

var supportOS = map[string]struct{}{
	"linux": {},
}

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Use config CLI to configure proxy settings.",
		Run: func(cmd *cobra.Command, args []string) {
			set, _ := cmd.Flags().GetString("set")
			config(set)
		},
	}

	cmd.Flags().StringP("set", "s", "0", `
--set 0 -> use 0 to clear proxy config
--set 127.0.0.1:8080 -> with domain param to set proxy config
`)

	return cmd
}

// config 代理配置
//
// set: 0 , 清除配置的代理
// set: 127.0.0.1:8080, 设置代理地址
func config(set string) {
	if _, ok := supportOS[runtime.GOOS]; !ok {
		println("auto config not support for os:" + runtime.GOOS)
		return
	}
	if set == "0" {
		configClear()
	} else {
		configClear()
		configSet(set)
	}
}

func configClear() {
	if err := removeFromEnv("http_proxy"); err != nil {
		println("proxy config clear failed", err.Error())
		return
	}
	if err := removeFromEnv("https_proxy"); err != nil {
		println("proxy config clear failed", err.Error())
		return
	}
	println("proxy config clear successfully")
}

func configSet(domain string) {
	if err := appendToEnv("http_proxy", domain); err != nil {
		println("proxy config set failed", err.Error())
		return
	}
	if err := appendToEnv("https_proxy", domain); err != nil {
		println("proxy config set failed", err.Error())
		return
	}
	println("proxy config set successfully")
}

func appendToEnv(variable, value string) error {
	file, err := os.OpenFile("/etc/environment", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("\n%s=\"%s\"", variable, value))
	return err
}

func removeFromEnv(variable string) error {
	envPath := "/etc/environment"
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", envPath, err)
	}
	lines := strings.Split(string(content), "\n")

	var newLines []string
	for _, line := range lines {
		if strings.Contains(line, variable+"=") {
			continue
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(envPath, []byte(newContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", envPath, err)
	}

	return nil
}
