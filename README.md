# memssh

memssh is a secure, portable command-line SSH client written in Go. It simplifies connecting to remote SSH servers using in-memory private keys and maintains host fingerprints in a user-local known_hosts.json file.


## Why memssh?

memssh was created to safely use SSH private keys that are stored in **untrusted locations** such as:

- USB drives
- Encrypted or ephemeral file systems
- Crypto containers
- Temporary key provisioning environments

With memssh, your private key can be loaded **entirely into memory** without ever being written to disk. This makes it ideal for high-security workflows where key persistence is unacceptable.


## Features

- Lightweight SSH client
- Supports in-memory private key authentication
- Trusted host fingerprint validation with prompt
- Stores fingerprints in ~/.ssh/known_hosts.json (cross-platform)
- Interactive shell or remote command execution
- Optional storage bypass (-no-store)
- Securely wipes key/passphrase memory after use


## Platforms

- Supported on Windows (cmd or PowerShell)
- Supported on Linux
- Supported on macOS


## Installation

### Requirements

- Go 1.20 or newer

### Get Dependencies

Run the following to get the required packages:

```bash
go get golang.org/x/crypto/ssh
go get golang.org/x/term
```

### Build (Linux/macOS)

```bash
git clone https://github.com/yourusername/memssh.git
cd memssh
go build -o memssh main.go
```

### Build (Windows)

```powershell
git clone https://github.com/yourusername/memssh.git
cd memssh
go build -o memssh.exe main.go
```


## Usage

### Basic Example (Using Private Key File)

```bash
memssh -host 192.168.1.10 -user ubuntu -key /path/to/id_rsa
```

### Run a Single Command

```bash
memssh -host server.example.com -user admin -key ~/.ssh/id_ed25519 -cmd "uptime"
```

### Paste Private Key at Runtime (No -key flag)

If you omit the -key flag, you will be prompted to paste your private key directly into the terminal:

```bash
memssh -host 192.168.1.10 -user ubuntu
Paste your private key (end with an empty line):
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

### Skip Saving Host Fingerprints

You can prevent memssh from saving the host fingerprint locally using -no-store:

```bash
memssh -host test.server.local -user dev -key ./temp_key.pem -no-store
```


## Known Hosts Storage

Trusted fingerprints are stored in:

- Linux/macOS: $HOME/.ssh/known_hosts.json
- Windows: %USERPROFILE%\.ssh\known_hosts.json

You can manually edit this file to remove or inspect fingerprints.


## Security Considerations

memssh was designed with security in mind, particularly for sensitive workflows involving untrusted or temporary storage.

Key security features:

- **In-memory private keys**: Private keys can be supplied via `stdin` and are never written to disk, allowing safe use from USB drives, encrypted containers, or ephemeral environments.
- **Memory wiping**: Key material and passphrases are explicitly zeroed from memory after use to reduce exposure.
- **Host fingerprint verification**: Server fingerprints are validated using SHA-256 hashes and stored in a local known_hosts database with user confirmation.
- **User-controlled trust**: On fingerprint change or first connection, the user must explicitly confirm trust, preventing silent man-in-the-middle acceptance.
- **No background daemons**: memssh is a single-run utility that exits cleanly after session or command execution.
- **Cross-platform path security**: Known hosts are stored securely under `$HOME/.ssh` or `%USERPROFILE%\.ssh`, consistent with OpenSSH best practices.
- **No key agent exposure**: This tool does not interface with SSH agents or forward credentials, ensuring isolation of authentication material.

> While memssh uses cryptographically secure libraries and follows best practices, it is still a CLI tool and should be used responsibly. Always review code and dependencies in high-security deployments.


## License

memssh is licensed under the Mozilla Public License 2.0 (MPL-2.0).

You may use, modify, and redistribute this software under the terms of the MPL. Contributions are welcome.

See LICENSE at https://www.mozilla.org/en-US/MPL/2.0/ for full terms.


## Third-Party Modules

memssh uses the following open-source modules:

- golang.org/x/crypto/ssh – SSH client implementation  
  License: BSD-3-Clause
- golang.org/x/term – Secure password and terminal handling  
  License: BSD-3-Clause


## Contributing

Feel free to open issues or submit pull requests for enhancements or bugfixes. All contributions must be MPL-2.0 compatible.


## Author

Copyright 2025 ffarkas  
Licensed under the Mozilla Public License 2.0
