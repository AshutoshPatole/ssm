![ssm-logo](./ssm.webp)

# SSM (Simple SSH Manager)

SSM is a versatile command-line tool for managing SSH connections and user authentication. It simplifies the management of SSH profiles with commands to register users, import configurations, connect to remote servers, and synchronize settings across devices.

## Features

- Add and manage SSH server configurations
- Connect to servers using stored profiles
- Import configurations from YAML files
- Generate template configuration files
- Sync configurations across devices using Firebase
- Download files from remote servers
- Update the CLI tool to the latest version

## Installation

Check [Release](https://github.com/AshutoshPatole/ssm-v2/releases) section for binaries

## Usage

### Basic Commands

- `ssm add`: Add a new SSH server configuration
- `ssm connect`: Connect to a server
- `ssm delete`: Delete a server configuration
- `ssm import`: Import configurations from a YAML file
- `ssm template`: Generate a template configuration file
- `ssm update`: Check for and install updates
- `ssm rdp`: Connect to windows machine using RDP.

### Authentication and Sync

- `ssm auth register`: Register a new user
- `ssm sync push`: Upload configurations to the cloud
- `ssm sync pull`: Download configurations from the cloud

### File Operations

- `ssm reverse-copy`: Download files from a remote server (TUI)

### Additional Features

- `ssm rdp`: Connect to Windows servers using RDP (Linux only)

## Configuration

SSM uses a YAML configuration file located at `$HOME/.ssm.yaml`. You can specify a different configuration file using the `--config` flag.

## Examples

1. Add a new server:
   ```
   ssm add hostname -u username -g group -e environment -a alias
   ```

2. Connect to a server:
   ```
   ssm connect group-name
   ```

3. Import configurations:
   ```
   ssm import -f config.yaml -g group-name
   ```

4. Sync configurations:
   ```
   ssm sync push -e user@example.com
   ssm sync pull -e user@example.com
   ```

5. Download files from a remote server:
   ```
   ssm reverse-copy group-name
   ```
6. Connect to windows using RDP
    ```
    ssm rdp
    ```
