# SSM - Simple SSH Manager

<p align="center">
  <img src="ssm.webp" alt="SSM Logo" width="300"/>
</p>

SSM (Simple SSH Manager) is a versatile command-line tool designed to streamline the management of SSH connections and user authentication.

Why SSM?
- Simplifies management of multiple SSH profiles
- Enhances security with built-in key rotation and encryption
- Saves time with quick connect commands and configuration imports
- Supports both SSH and RDP connections for comprehensive server management

How SSM helps:
1. Centralized Configuration: Store all your SSH profiles in one secure, easy-to-manage location.
2. Quick Connections: Connect to any server with a single command, eliminating the need to remember IP addresses and usernames.
3. Enhanced Security: Regularly rotate SSH keys and encrypt sensitive data to maintain robust security practices.


Whether you're managing a handful of servers or a large-scale infrastructure, SSM simplifies your workflow, enhances security, and saves valuable time in your day-to-day operations.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Commands](#commands)
  - [User Management](#user-management)
  - [Server Management](#server-management)
  - [Connection](#connection)
  - [Synchronization](#synchronization)
  - [Utilities](#utilities)

## Installation
To install SSM, download the latest release from the [GitHub Releases page](https://github.com/AshutoshPatole/ssm/releases).

Choose the appropriate version for your operating system and architecture. For example, on Linux:

```bash
wget https://github.com/AshutoshPatole/ssm/releases/download/v0.2.0/ssm_Linux_x86_64.tar.gz
tar xzf ssm_Linux_x86_64.tar.gz
sudo mv ssm /usr/local/bin/
```

Replace `v.2.0` with the latest version number and adjust the filename according to your system.

## Configuration

SSM uses a YAML configuration file located at `~/.ssm.yaml`. You can generate a template configuration using the `template` command. Here's a sample YAML configuration:

```yaml
groups:
  - name: production
    environment:
      - name: prod
        servers:
          - hostname: prod-server-1.example.com
            alias: prod1
            user: admin
          - hostname: prod-server-2.example.com
            alias: prod2
            user: root
  - name: development
    environment:
      - name: dev
        servers:
          - hostname: dev-server.example.com
            alias: dev1
            user: developer
      - name: staging
        servers:
          - hostname: staging-server.example.com
            alias: staging1
            user: tester
firebaseConfig: /path/to/downloaded/service-account.json
firebaseApiKey: YOUR_FIREBASE_API_KEY
```

This configuration defines two groups (production and development) with different environments and servers. You can customize this structure to fit your specific needs.

The `firebaseConfig` and `firebaseApiKey` fields are used to configure your own Firebase cloud for synchronization:

- `firebaseConfig`: The path to your Firebase service account JSON file. This file contains the credentials needed to authenticate your application with Firebase.
- `firebaseApiKey`: Your Firebase API key, which is used for client-side authentication.

To obtain these values:

1. Go to the [Firebase Console](https://console.firebase.google.com/)
2. Create a new project or select an existing one
3. Go to Project Settings > Service Accounts
4. Click "Generate new private key" to download the service account JSON file
5. Set the `firebaseConfig` path to the location where you saved this file
6. Go to Project Settings > General
7. Copy the "Web API Key" and set it as the `firebaseApiKey` value

By providing these configurations, you can use your own Firebase project for syncing SSM data, ensuring that your sensitive information is stored in your own cloud environment.

## Commands

### User Management

#### Register

Register a new user for SSM:

```bash
ssm auth register --email user@example.com
```

This command registers a new user with the provided email address. It will prompt for a password securely.

#### Reset Password

Reset the password for an existing user:

```bash
ssm auth reset-password --email user@example.com
```

This command initiates the password reset process for the specified email address. A password reset link will be sent to the user's email.

### Server Management

#### Add

Add a new SSH server configuration:

```bash
ssm add example.com --username root --group production --alias prod-server --environment prod
```

This command adds a new server configuration. It will prompt for the server's password if not provided.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --username, -u | Username to use | root |
| --group, -g | Group to use | (required) |
| --alias, -a | Alias for the server | (required) |
| --environment, -e | Environment to use | dev |
| --rdp, -r | Flag to indicate it's an RDP connection | false |

#### Delete

Remove a server configuration:

```bash
ssm delete --server prod-server
```

This command removes a server configuration from SSM.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --server, -s | Server to delete | (required) |
| --clean-config, -c | Clean unused groups | false |

#### Import

Import SSH configurations from a YAML file:

```bash
ssm import --file config.yaml --group production
```

This command imports SSH configurations from a specified YAML file.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --file, -f | File path | (required) |
| --group, -g | Group name | "" |
| --all, -a | Import all groups | false |
| --setup-dot | Setup dot files in servers | false |

### Connection

#### Connect

Connect to a server:

```bash
ssm connect production
```

This command connects to a server in the specified group.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --filter, -f | Filter list by environment | "" |

#### RDP

Connect to a Windows server using RDP:

```bash
ssm rdp production --filter dev
```

This command connects to a Windows server using RDP.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --filter, -f | Filter list by environment | "" |

### Synchronization

#### Push

Upload your configuration and sensitive files to the cloud:

```bash
ssm sync push --email user@example.com
```

This command securely uploads your local SSM configuration, SSH keys, and dotfiles to the cloud. The following files are uploaded:

1. `.ssm.yaml`: Your SSM configuration file
2. `.ssh/id_ed25519`: Your SSH private key
3. `.ssh/id_ed25519.pub`: Your SSH public key
4. `.zshrc`: Your Zsh configuration (if present)
5. `.bashrc`: Your Bash configuration (if present)
6. `.tmux.conf`: Your Tmux configuration (if present)
7. `.ssh/config`: Your SSH client configuration

All files are encrypted before upload using AES-256 encryption. The encryption key is derived from your password using PBKDF2 with SHA-256. Encrypted data is stored in Firebase, ensuring secure cloud storage.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --email, -e | Email address | (required) |

#### Pull

Download your configuration from the cloud:

```bash
ssm sync pull --email user@example.com
```

This command downloads your SSM configuration from the cloud.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --email, -e | Email address | (required) |

### Utilities

#### Rotate Key

Rotate SSH keys for added security:

```bash
ssm rotate-key --all --private-key ~/.ssh/id_ed25519 --public-key ~/.ssh/id_ed25519.pub
```

This command rotates SSH keys for all or a specific group of servers.

| Argument | Description | Default Value |
|----------|-------------|---------------|
| --all | Rotate keys for all servers | false |
| --group | Rotate keys for a specific group | "" |
| --private-key | Path to the Ed25519 private key | (required) |
| --public-key | Path to the Ed25519 public key | (required) |

#### Template

Generate a template YAML configuration file:

```bash
ssm template
```

This command generates a template YAML configuration file and saves it as `.ssm-template.yaml` in the user's home directory.

#### Update

Check for and install updates:

```bash
ssm update
```

This command checks for available updates and installs them if found.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
