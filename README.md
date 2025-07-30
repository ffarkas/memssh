
# memssh

`memssh` is a lightweight SSH client wrapper built to address a specific issue: **OpenSSH blocks SSH keys stored in insecure locations** (such as world-readable folders or custom password managers). This tool allows interactive or command-based SSH access using a pasted private key, bypassing strict file permission checks while still performing secure host verification and fingerprint pinning.

## ✨ Features

- Connects to SSH servers using a private key passed via argument or stdin
- Secure fingerprint verification with user confirmation
- Stores known hosts and their fingerprints in a local `.known_hosts.json` file
- Supports command execution or full interactive shells
- Handles encrypted private keys with passphrase prompts
- Works well in restricted environments (e.g., when `.ssh/` cannot be used)

## 🔧 Why It Exists

Standard `ssh` rejects keys that aren’t stored securely on disk. This can be frustrating in environments where:

- You store keys in secure password managers or hardware tokens
- You can’t or don’t want to use the default `~/.ssh` directory
- You run in restricted containers, CI/CD environments, or minimal systems

`memssh` solves this by allowing you to paste a key or pass it as a flag, with secure handling and ephemeral memory wiping after use.

## 🚀 Usage

### Install dependencies

Make sure you have Go installed (version 1.18 or later recommended). Then fetch the dependencies:

```bash
go get golang.org/x/crypto/ssh
go get golang.org/x/term
```

### Build

```bash
go build -o memssh
```

### Run

```bash
./memssh -host example.com -user ubuntu
```

Or with a pasted key:

```bash
./memssh -host example.com -user ubuntu
Paste your private key (end with an empty line):
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

You can also pass the key via `-key` flag:

```bash
./memssh -host example.com -user ubuntu -key "/path/to/key"
```

To run a command instead of an interactive shell:

```bash
./memssh -host example.com -user ubuntu -cmd "uptime"
```

Disable saving of host fingerprints with:

```bash
./memssh -host example.com -user ubuntu -no-store
```

## 🗂 Known Hosts Storage

Fingerprint trust is stored locally in `.known_hosts.json` in the **same directory as the compiled binary**. This file maps `host:port` → `SHA256 fingerprint`. If a fingerprint changes, you're prompted to accept the new one.

## 🔐 Security Notes

- Keys are zeroed from memory after use.
- Passphrases are not stored.
- Only SHA256 fingerprints are used for verification.
- Interactive trust confirmation is required for unknown or changed hosts.

## 📦 Dependencies & Licenses

This tool uses the following Go packages:

- [`golang.org/x/crypto/ssh`](https://pkg.go.dev/golang.org/x/crypto/ssh)  
  Licensed under **BSD-style license** (compatible with open-source use)

- [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term)  
  Also licensed under the **BSD-style license**

These licenses do **not** require you to open-source your own code if you distribute a compiled binary, but **you must include the original license texts** if redistributing the source or as part of a bundled package. You can add a `LICENSES/` directory if needed.

## 📁 Example `.known_hosts.json`

```json
{
  "example.com:22": "abc123...base64-encoded-sha256..."
}
```

## 📃 License

This project is open-source under the [MIT License](LICENSE). See included third-party licenses if you distribute with dependencies.
