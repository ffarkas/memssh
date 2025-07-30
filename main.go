package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// KnownHosts maps SSH server addresses to their trusted public key fingerprints.
type KnownHosts map[string]string

func main() {
	// Define and parse command-line flags
	host := flag.String("host", "", "SSH server hostname or IP")
	port := flag.Int("port", 22, "SSH server port")
	user := flag.String("user", "", "SSH username")
	key := flag.String("key", "", "SSH private key (PEM format) (optional)")
	cmd := flag.String("cmd", "", "Command to run on remote server (optional)")
	noStore := flag.Bool("no-store", false, "Do not store new or changed host fingerprints")
	flag.Parse()

	if *host == "" || *user == "" {
		flag.Usage()
		log.Fatal("host and user are required")
	}

	privateKey := getPrivateKey(*key)
	defer zeroBytes(privateKey)

	signer, err := parsePrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Private key error: %v", err)
	}

	address := fmt.Sprintf("%s:%d", *host, *port)
	knownHostsPath := getKnownHostsPath()
	knownHosts := loadKnownHosts(knownHostsPath)

	config := &ssh.ClientConfig{
		User:            *user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback(address, knownHosts, knownHostsPath, *noStore),
	}

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	if *cmd == "" {
		startInteractiveShell(client)
	} else {
		runCommand(client, *cmd)
	}
}

// getPrivateKey loads a private key from a file path or inline input.
// If the `pathOrInline` is empty, it prompts the user for multiline pasted key input.
func getPrivateKey(pathOrInline string) []byte {
	if pathOrInline == "" {
		fmt.Print("Paste your private key (end with an empty line):\n")
		data, err := readMultiLineInput()
		if err != nil {
			log.Fatalf("Failed to read private key: %v", err)
		}
		return data
	}
	if data, err := os.ReadFile(pathOrInline); err == nil {
		return data
	}
	// Fallback: treat input as inline PEM key
	return []byte(pathOrInline)
}

// parsePrivateKey attempts to parse an SSH signer from a PEM private key.
// If the key is encrypted, it prompts the user for the passphrase.
func parsePrivateKey(key []byte) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err == nil {
		return signer, nil
	}
	if !strings.Contains(err.Error(), "encrypted") {
		return nil, err
	}

	fmt.Print("Enter passphrase for encrypted private key: ")
	pass, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("reading passphrase failed: %w", err)
	}
	defer zeroBytes(pass)

	return ssh.ParsePrivateKeyWithPassphrase(key, pass)
}

// hostKeyCallback returns an ssh.HostKeyCallback that checks a known_hosts map
// for matching fingerprints and optionally prompts to trust and save new or changed ones.
func hostKeyCallback(address string, known KnownHosts, path string, noStore bool) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		hash := sha256.Sum256(key.Marshal())
		fp := base64.StdEncoding.EncodeToString(hash[:])

		if stored, exists := known[address]; exists {
			if stored == fp {
				return nil
			}
			fmt.Printf("\nWARNING: fingerprint for %s has changed!\nOld: %s\nNew: %s\n", address, stored, fp)
			fmt.Print("Do you want to overwrite and trust the new fingerprint? (y/n): ")
			if !askYesNo() {
				return fmt.Errorf("fingerprint mismatch rejected by user")
			}
		} else {
			fmt.Printf("\nNew host: %s\nFingerprint: %s\nTrust this host? (y/n): ", address, fp)
			if !askYesNo() {
				return fmt.Errorf("user declined to trust unknown host")
			}
		}

		if !noStore {
			known[address] = fp
			saveKnownHosts(path, known)
			fmt.Println("Host fingerprint saved.")
		} else {
			fmt.Println("Fingerprint not saved due to -no-store flag.")
		}
		return nil
	}
}

// runCommand runs a remote command on the SSH server and prints its output.
func runCommand(client *ssh.Client, cmd string) {
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	fmt.Printf("Running command: %s\n", cmd)
	if err := session.Run(cmd); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

// startInteractiveShell starts a full interactive terminal session on the remote SSH server.
func startInteractiveShell(client *ssh.Client) {
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	fd := int(syscall.Stdin)
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		log.Fatalf("Failed to set terminal raw mode: %v", err)
	}
	defer term.Restore(fd, oldState)

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	width, height, _ := term.GetSize(fd)
	if width == 0 || height == 0 {
		width, height = 80, 24
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", height, width, modes); err != nil {
		log.Fatalf("PTY request failed: %v", err)
	}

	go handleSignals(session)

	if err := session.Shell(); err != nil {
		log.Fatalf("Failed to start shell: %v", err)
	}
	if err := session.Wait(); err != nil {
		log.Fatalf("Shell exited with error: %v", err)
	}
}

// getKnownHostsPath returns the path to the local known_hosts.json file in ~/.ssh.
func getKnownHostsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to determine user home directory: %v", err)
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		log.Fatalf("Failed to create .ssh directory: %v", err)
	}
	return filepath.Join(sshDir, "known_hosts.json")
}

// loadKnownHosts loads the known_hosts.json file into memory, or returns an empty map if not found.
func loadKnownHosts(path string) KnownHosts {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return KnownHosts{}
		}
		log.Fatalf("Failed to open known_hosts: %v", err)
	}
	defer file.Close()

	var hosts KnownHosts
	if err := json.NewDecoder(file).Decode(&hosts); err != nil {
		log.Printf("Warning: could not parse known_hosts.json: %v", err)
		return KnownHosts{}
	}
	return hosts
}

// saveKnownHosts writes the updated known hosts map to known_hosts.json.
func saveKnownHosts(path string, hosts KnownHosts) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to write known_hosts: %v", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(hosts); err != nil {
		log.Fatalf("Failed to encode known_hosts: %v", err)
	}
}

// askYesNo prompts the user for a yes/no answer and returns true if the answer begins with "y" or "Y".
func askYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(input)), "y")
}

// readMultiLineInput reads lines from stdin until an empty line is encountered.
// Used for pasting multi-line private keys.
func readMultiLineInput() ([]byte, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	return []byte(strings.Join(lines, "\n")), scanner.Err()
}

// zeroBytes overwrites a byte slice with zeroes to securely erase sensitive data like private keys or passphrases.
func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
