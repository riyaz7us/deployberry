# DeployBerry - Simple Application & Server Control Panel

DeployBerry is an easy-to-use control panel designed to help you host and manage websites, databases, and apps on your server without needing complex shell scripting. Whether you are running a simple **WordPress blog**, an **Odoo business suite**, or running custom container apps, DeployBerry handles the server setup (Nginx, Caddy, databases, cron jobs, and processes) automatically using straightforward configurations.

---

## 🎯 Features

- **No Dev Skills Required**: Launch database servers, PHP configurations, and web proxies from an intuitive web dashboard.
- **Unified & Pluggable**: Run as a complete server panel or compile it as a standalone container controller.
- **Easy Configuration**: Edit application configuration files (like `.env` or settings files) directly in your browser.
- **Dynamic Action Buttons**: Add custom controls (like clearing cache, importing databases) that show up directly as dashboard buttons.
- **Container Deployments**: Deploy Podman compose templates or custom Docker images in seconds.

---

## 🚀 Quick Install

To install DeployBerry on your Linux server, run the quick install script as root:

```bash
curl -fsSL https://raw.githubusercontent.com/riyaz7us/deployberry/master/install.sh | sudo bash
```

This script will automatically:
1. Set up the panel directories and isolated environment.
2. Install the `deployberry` executable.
3. Register the systemd service to keep the panel running in the background.
4. Prompt you to register your admin user.

---

## 💻 CLI Commands

Once installed, you can use the `deployberry` CLI command in your terminal to manage the panel:

| Command | Description | Example |
| :--- | :--- | :--- |
| `install` | Sets up system paths, directories, and background services. | `sudo deployberry install` |
| `register` | Registers a new panel administrator account. | `sudo deployberry register admin mypassword` |
| `status` | Checks if the background panel service is active. | `deployberry status` |
| `start` / `stop` / `restart` | Restarts, starts, or stops the panel daemon. | `sudo deployberry restart` |
| `logs` | Streams the real-time runtime log output of the panel. | `deployberry logs` |
| `exec` | Safely executes commands inside the isolated app user environment. | `deployberry exec composer install` |
| `uninstall` | Completely removes DeployBerry binaries and configs. | `sudo deployberry uninstall` |

---

## 🛠️ Usage

### Logging In
Once the installation script runs successfully, access your panel at:
`http://your-server-ip:7717`

### Hosting Applications & Websites
1. Head over to the **Registry** page to pick from pre-packaged templates (WordPress, Node, static sites).
2. Enter your custom domain name, select your database engine, specify environment details, and click **Install**.
3. Manage domain routes, databases, reverse proxy configurations (Nginx/Caddy), PM2 processes, and logs from the **Applications** dashboard.

---

## 📖 Developer Documentation & Manual Builds

If you want to manually build DeployBerry from source, configure modular build tags, or inspect the REST API endpoints, please check:
👉 **[developer.md](developer.md)**

---

## 📋 Implementation Roadmap

### **Phase 1: Core System (COMPLETED)**
- [x] Manifest-driven generic server panel and application installer.
- [x] Web-based configuration file editor.
- [x] Podman Compose and custom container orchestration.
- [x] Standalone compilation mode (`-tags containerapps_only`).

### **Phase 2: Service Enhancements (IN PROGRESS)**
- [ ] Automated background backup scheduling and retention quotas.
- [ ] Log buffering, rotation, and real-time console streaming.
- [ ] Cron manager scheduling controls and system log visualizer.
- [ ] Active service monitors for system metrics.

---

## 🤝 Contributing

We welcome additions of new application manifests and panel feature expansions:
1. Fork this repository.
2. Define template manifests in the [deployberry-manifests](https://github.com/riyaz7us/deployberry-manifests) repository.
3. Submit a pull request.

---

## 🏆 Goal

Our mission is to create a panel that is **generic by design**. We strive to move all custom software logic out of the Go code and into configuration manifests. This ensures that any user—from business owners to system administrators—can deploy and manage any service configuration without code updates. 🚀
