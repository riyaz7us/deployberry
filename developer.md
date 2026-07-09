# Developer Documentation & API Reference

This document outlines the architecture, layout, and API endpoints implemented in the DeployBerry workspace.

---

## 🏗️ Architecture Layout

The codebase is organized as a Go workspace (`go.work`) with decoupled modules:

1.  **`shared`**: Core shared utilities (database setup, shell commands, file utilities, and the Nginx/Caddy web server config orchestrators).
2.  **`containerapps`**: Podman Compose orchestration module. Completely decoupled from the panel frontend logic.
3.  **`deployberry`**: The primary compilation module. Includes the modular route registration engine (`mod_core.go` and `mod_containerapps.go`) supporting conditional compiling.
4.  **`ui`**: Responsive frontend built in Nuxt.js.

---

## 🛠️ Compilation & Manual Builds

DeployBerry supports conditional compilation via Go workspace build tags:

### 1. Build Complete Unified Panel
Compiles the core engine, standard dashboard modules (cron, PM2, DB interfaces), UI assets, and containerapps registry:
```bash
# Option 1: Run via Makefile
make backend

# Option 2: Build directly
cd deployberry && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o ../bin/deployberry-linux-amd64 .
```

### 2. Build Standalone ContainerApps Daemon
Compiles only the containerapps routes and basic daemon, pruning all standard dashboard modules (database accounts, crons, files, process monitors, and web UI serving assets):
```bash
# Option 1: Run via Makefile
make containerapps

# Option 2: Build directly
cd deployberry && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags containerapps_only -buildvcs=false -o ../bin/containerapps-linux-amd64 .
```

### 3. Running in Development
To run the server in development without compiling to a binary:

**Option 1: Using the root Makefile (Recommended)**
```bash
# To run the complete unified panel
make run

# To run only the ContainerApps daemon
make run-containerapps
```

**Option 2: Running directly via Go CLI**
```bash
# Navigate to the main project directory
cd deployberry

# Run the complete unified panel
go run .

# Run only the ContainerApps daemon
go run -tags containerapps_only .
```

---

## 🔌 API Documentation

All API endpoints are prefixed with `/api` and require authorization (using the `AuthMiddleware`) unless specified as public.

### 🔐 Authentication

#### `POST /api/login` (Public)
Authenticates a user and starts a session.
- **Request Body**:
  ```json
  {
    "username": "admin",
    "password": "password"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "token": "JWT_TOKEN_STRING"
  }
  ```

#### `GET /api/logged-in`
Verifies authorization session validity.
- **Response**: `200 OK` (with authorization confirmation).

---

### 🌐 Webserver Configurations

These endpoints manage Nginx, Caddy, and unified web configuration database records.

#### `GET /api/webconfigs`
Lists all generated web configs stored in the database.
- **Response**:
  ```json
  {
    "success": true,
    "configs": [
      {
        "domain": "example.com",
        "root_path": "/var/www/html/example.com",
        "php_version": "8.2",
        "reverse_proxy_url": "",
        "enable_gzip": true,
        "enable_cache": true,
        "webserver": "nginx",
        "ssl": false,
        "created_at": "2026-07-03T16:31:41Z",
        "updated_at": "2026-07-03T16:31:41Z"
      }
    ]
  }
  ```

#### `POST /api/webconfigs`
Creates a new webserver configuration (generating both Nginx and Caddy config files and saving details to DB).
- **Request Body**:
  ```json
  {
    "domain": "example.com",
    "root_path": "/var/www/html",
    "php_version": "8.2",
    "reverse_proxy_url": "",
    "enable_gzip": true,
    "enable_cache": false
  }
  ```

#### `POST /api/webconfigs/:domain/deploy`
Force deploys an existing configuration to the active system folders.
- **Query Parameter**: `?server=nginx|caddy|both` (default is `both`)

#### `POST /api/webconfigs/:domain/recreate`
Cleans up and recreates configuration files, then redeploys.

#### `DELETE /api/webconfigs/:domain`
Removes configuration files from system paths and deletes GORM database records.

---

### 📦 Applications Manager (Unified Lifecycle)

These endpoints handle both native panel apps and containerized (Podman) apps uniformly.

#### `GET /api/applications`
Queries all installed applications.
- **Query Parameters**: `?search=`, `?status=`, `?provider=`
- **Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 1,
        "path": "/opt/apps/my-app",
        "domain": "my-app.localhost",
        "provider": "custom",
        "title": "my-app",
        "display_name": "My App",
        "version": "1.0",
        "runtime": "podman",
        "database": "",
        "status": "running"
      }
    ]
  }
  ```

#### `GET /api/applications/:id`
Retrieves application details, configuration properties, and its manifest-defined commands.

#### `GET /api/applications/:id/status`
Checks runtime execution status, resources (CPU/memory), and process metrics.

#### `POST /api/applications/:id/start`
Starts the application (PM2 process or Podman stack).

#### `POST /api/applications/:id/stop`
Stops the application.

#### `POST /api/applications/:id/restart`
Restarts the application.

#### `POST /api/applications/:id/update`
Runs update steps defined in the manifest.

#### `POST /api/applications/:id/command`
Runs a custom manifest command.
- **Request Body**:
  ```json
  {
    "command": "build_assets",
    "args": {
      "ENV": "production"
    }
  }
  ```

#### `GET /api/applications/:id/logs`
Fetches standard outputs/errors of the running application.
- **Query Parameter**: `?lines=100`

#### `GET /api/applications/:id/files`
Lists editable configuration files configured by the manifest.

#### `DELETE /api/applications/:id`
Gracefully stops execution, cleans up databases/webservers, deletes folder content (optional), and drops DB records.

---

### 📥 Application Installer (Registry)

Declarative application installer for deploying applications on the host environment.

#### `GET /api/registry`
Lists applications available in the registry index.

#### `GET /api/registry/:slug/requirements`
Fetches the manifest for a registry package and determines compatible languages/databases.

#### `POST /api/registry/:slug/install`
Triggers manifest installation. Builds folders, configures local DB credentials, executes installation scripts, and deploys Nginx/Caddy proxy.
- **Request Body**:
  ```json
  {
    "appName": "test-wordpress",
    "domain": "test.localhost",
    "path": "/opt/deployberry/apps",
    "deploymentMethod": "git",
    "gitRepo": "https://github.com/WordPress/WordPress.git",
    "databaseEngine": "mysql",
    "vars": {}
  }
  ```

---

### 🐳 Container Apps Registry

Manages declarations and installations for Podman containers.

#### `GET /api/containerapps/registry`
Lists container app templates in the registry.

#### `GET /api/containerapps/registry/:slug/requirements`
Checks requirements for container deployment and checks if Podman/Compose is installed.

#### `POST /api/containerapps/registry/:slug/install`
Deploys compose template or docker image, binds host ports, and deploys reverse proxy configs.
- **Request Body**:
  ```json
  {
    "appName": "my-nginx",
    "domain": "nginx.localhost",
    "path": "/opt/apps",
    "deploymentMethod": "none",
    "image": "docker.io/library/nginx:alpine",
    "container_port": 80
  }
  ```

---

## 📜 App Manifest Schema & System Variables

Application manifests are YAML files defined inside [deployberry-manifests](https://github.com/riyaz7us/deployberry-manifests) which describe runtime, package requirements, build/installation steps, and process configurations.

The master reference for the manifest format and its keys is:
- [manifest.schema.yaml](https://github.com/riyaz7us/deployberry-manifests/blob/master/manifest.schema.yaml)

### System Variables Reference

During installation, updating, or running actions, several variables are automatically populated by DeployBerry and made available to step command scripts (via `{VARIABLE_NAME}` substitution):

| Variable | Description |
|---|---|
| `{APP_PATH}` | Absolute filesystem installation path of the application. |
| `{APP_NAME}` | Unique lowercase slug/name of the application. |
| `{DOMAIN}` | Primary domain configured for the application. |
| `{APP_PORT}` | Local port override. Configured via Nginx/Caddy proxy and available to bind the app server. |
| `{DB_HOST}` | Database server hostname (typically `localhost`). |
| `{DB_NAME}` | Generated database name (if database is required). |
| `{DB_USER}` | Generated database username (if database is required). |
| `{DB_PASS}` | Generated database password (if database is required). |
| `{DB_PORT}` | Database port. |
| `{DEPLOYMENT_METHOD}` | The user-selected source deployment method (`git`, `upload`, `manual`). |
| `{GIT_REPO}` | Source Git URL (populated when using `git` deployment method). |
| `{GIT_BRANCH}` | Configured Git branch to checkout (populated when using `git` deployment method). |
| `{MANUAL_PATH}` | Direct local path target (populated when using `manual` deployment method). |
| `{ARGS}` | Dynamic user-supplied arguments (available when executing manifest commands). |

### Brief Instruction for Writing Manifests

1. **Use matching file names**: Place your manifest in the registry directory [deployberry-manifests](https://github.com/riyaz7us/deployberry-manifests) named as `<slug>.manifest.yaml` where `<slug>` matches the `name` field in the manifest.
2. **Register the Manifest**: Append a corresponding JSON object with matching `slug` to [index.json](https://github.com/riyaz7us/deployberry-manifests/blob/master/index.json).
3. **Configure Custom Ports**: For proxy applications (e.g. Node, Python), include the `APP_PORT` key in variables. This permits user-configurable routing, which overrides the webserver proxy configuration and provides `{APP_PORT}` to your startup process scripts.

