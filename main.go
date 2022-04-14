package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	sshcfg "github.com/kevinburke/ssh_config"
	"github.com/manifoldco/promptui"
)

func main() {
	config := flag.String("config", ".ssh/config", "ssh config filename")
	flag.Parse()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("os.UserHomeDir: %v", err)
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", homeDir, *config))
	if err != nil {
		log.Fatalf("os.Open: %v", err)
	}

	defer file.Close()

	var hosts []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, "Host") {
			continue
		}

		cols := strings.Fields(line)

		if len(cols) < 2 {
			log.Fatalf("no host in line: %v", line)
		}

		hosts = append(hosts, cols[1])
	}

	var servers []Server
	for _, v := range hosts {
		servers = append(servers, Server{
			Host: v,
			User: sshcfg.Get(v, "User"),
		})
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "❤️  {{ .Host | cyan }}",
		Inactive: "  {{ .Host | cyan }}",
		Selected: "Excellent choice {{ .Host | green }}",
	}

	prompt := promptui.Select{
		Label:     "Choose Your Fighter",
		Items:     servers,
		Size:      10,
		Templates: templates,
		Searcher: func(input string, index int) bool {
			server := servers[index]

			host := strings.Replace(strings.ToLower(server.Host), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(host, input)
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	server := servers[i]

	binary, err := exec.LookPath("ssh")
	if err != nil {
		log.Fatalf("exec.LookPath: %v", err)
	}

	if err := syscall.Exec(
		binary,
		[]string{"ssh", fmt.Sprintf("%s@%s", server.User, server.Host)},
		os.Environ(),
	); err != nil {
		log.Fatalf("syscall.Exec: %v", err)
	}
}

type Server struct {
	Host string
	User string
}
