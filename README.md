# Envsgen

**A powerful configuration management tool that transforms TOML into multiple output formats**

Envsgen is a lightweight CLI utility that reads TOML configuration files, resolves variable interpolations, and generates environment configurations in various formats including dotenv, JSON, YAML, Docker Compose, Caddyfile, and Bash scripts.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Real World Example:

- [Tutorial](./TUTORIAL.md) - Real World Example

## 🌟 Key Features

- **Multi-format output**: Generate dotenv, JSON, YAML, Docker Compose, Caddyfile, or Bash export scripts
- **Smart variable interpolation**: Reference other TOML keys, environment variables, or shell commands
- **Modular configuration**: Import other TOML files to keep configurations DRY
- **Section-based generation**: Extract specific configuration sections for different environments
- **Hierarchical support**: Optionally expand nested sections with namespaced keys
- **Environment inheritance**: Child sections automatically inherit and override parent values

---

## 📋 Table of Contents

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage](#-usage)
- [Configuration Structure](#-configuration-structure)
- [Variable Interpolation](#-variable-interpolation)
- [Output Formats](#-output-formats)
- [Advanced Features](#-advanced-features)
- [Real-World Examples](#-real-world-examples)
- [Troubleshooting](#-troubleshooting)
- [Roadmap](#-roadmap)

---

## 🚀 Installation

### Prerequisites

- Go 1.18 or higher (1.20+ recommended)

### Using Make (Recommended)

**Install to /usr/local/bin:**
```bash
make install
```

**Install to custom directory:**
```bash
INSTALL_DIR=$HOME/.local/bin make install
```

**Build only:**
```bash
make build
# Output: ./builds/envsgen
```

**Platform-specific builds:**
```bash
make linux      # Linux x86_64
make windows    # Windows x86_64
make darwin     # macOS x86_64
make silicon    # macOS ARM64 (M-series)
make release    # Build all platforms
```

### Using Go

**Build manually:**
```bash
go mod tidy
go build -o envsgen .
```

**Install via go install:**
```bash
go install ./
```

**Manual installation:**
```bash
sudo cp envsgen /usr/local/bin
```

---

## ⚡ Quick Start

### 1. Create a configuration file

**config.toml:**
```toml
[shared]
JWT_SECRET = "supersecret"
API_HOST = "https://api.example.com"

[backend]
BASE_URL = "${shared.API_HOST}/v1"
JWT_SECRET = "${shared.JWT_SECRET}"
PORT = 8000

[backend.local]
BASE_URL = "http://localhost:8000/v1"
DEBUG = true
```

### 2. Generate environment files

```bash
# Production dotenv
envsgen config.toml backend -o .env

# Local development dotenv
envsgen config.toml backend.local -o .env.local

# JSON output
envsgen config.toml backend --json -o config.json

# YAML output
envsgen config.toml backend --yaml -o config.yaml
```

---

## 📖 Usage

### Basic Syntax

```bash
envsgen <path/to/config.toml> <section> [options]
```

### Command-Line Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--dotenv` | `-de` | Output dotenv format (default) |
| `--json` | `-j` | Output JSON format |
| `--yaml` | `-y` | Output YAML format |
| `--caddy` | `-cy` | Output Caddyfile format |
| `--docker` | `-d` | Output Docker Compose YAML |
| `--envs`, `--bash` | `-ev` | Output Bash export script |
| `--output PATH` | `-o` | Write to file instead of stdout |
| `--expand` | `-e` | Include nested sections recursively |
| `--allow-shell` | | Enable shell command execution |
| `--ignore-missing-vars` | `-iv` | Ignore unresolved variables |
| `--strict-vars-check` | `-sv` | Fail on unresolved variables (default) |
| `--verbose` | `-v` | Enable verbose output |
| `--help` | `-h` | Show help message |

### Important Notes

- The `<section>` argument is **required** and must refer to a TOML table (e.g., `backend`, `backend.local`)
- Single keys (e.g., `backend.local.BASE_URL`) are not supported as sections
- Child sections inherit parent values unless overridden

---

## 🏗️ Configuration Structure

### Basic Structure

```toml
[shared]
# Shared variables accessible to all sections
API_VERSION = "v1"
DOMAIN = "example.com"

[backend]
# Backend-specific configuration
BASE_URL = "https://${shared.DOMAIN}/api/${shared.API_VERSION}"
PORT = 8000

[backend.local]
# Local override (inherits from [backend])
BASE_URL = "http://localhost:8000/api/${shared.API_VERSION}"
DEBUG = true
```

### Section Inheritance

When selecting `backend.local` **without** `--expand`:
- Inherits all keys from `[backend]`
- Overrides inherited keys with its own values
- Only immediate keys are output (no nested tables)

When selecting `backend` **with** `--expand`:
- Includes all nested sections
- Keys are namespaced: `LOCAL__DEBUG=true`
- Creates hierarchical structure

---

## 🔗 Variable Interpolation

### Syntax: `${path.to.value}`

Envsgen supports three types of variable interpolation:

### 1. TOML Key References

Reference other keys using dot notation:

```toml
[shared]
HOST = "example.com"
PORT = 443

[api]
BASE_URL = "https://${shared.HOST}:${shared.PORT}"
# Resolves to: https://example.com:443
```

### 2. Environment Variables

Access OS environment variables using the `envs` prefix:

```toml
[secrets]
API_KEY = "${envs.SECRET_API_KEY}"
DATABASE_PASSWORD = "${envs.DB_PASSWORD}"
```

```bash
export SECRET_API_KEY="abc123"
export DB_PASSWORD="secure_pass"
envsgen config.toml secrets
```

### 3. Shell Commands

Execute shell commands (requires `--allow-shell` flag):

```toml
[build]
GIT_COMMIT = "${`git rev-parse --short HEAD`}"
BUILD_TIME = "${`date -u +%Y-%m-%dT%H:%M:%SZ`}"
NODE_VERSION = "${`node --version`}"
```

```bash
envsgen config.toml build --allow-shell
```

### Recursive Resolution

Variables are resolved recursively:

```toml
[config]
STAGE = "production"
ENV_NAME = "${config.STAGE}"
LOG_FILE = "/var/log/${config.ENV_NAME}.log"
# Resolves to: /var/log/production.log
```

### Error Handling

**By default (strict mode):**
- Unresolved variables cause immediate exit with error
- Missing TOML keys trigger errors
- Variables resolving to objects (not primitives) trigger errors

**With `--ignore-missing-vars`:**
- Unresolved variables remain as `${...}` literals
- Warnings printed to stderr (with `--verbose`)
- Processing continues

---

## 📤 Output Formats

### 1. Dotenv (Default)

**Command:**
**TOML Source:**

```toml
[frontend]
NEXTAUTH_SECRET= "${shared.JWT_SECRET}"
GOOGLE_ID = "${shared.GOOGLE_ID}"
GOOGLE_SECRET = "${shared.GOOGLE_SECRET}"
RESEND_API_KEY = "${shared.RESEND_API_KEY}"

[frontend.local]
NEXTAUTH_URL = "http://localhost:3000"
MONGODB_URI = "${shared.mongodb.local.MONGODB_URI}"
STRIPE_PUBLIC_KEY = "${shared.stripe.test.STRIPE_PUBLIC_KEY}"
STRIPE_SECRET_KEY = "${shared.stripe.test.STRIPE_SECRET_KEY}"
STRIPE_WEBHOOK_SECRET = "${shared.stripe.test.STRIPE_WEBHOOK_SECRET}"
MY_NUMBER = "${shared.TEST_VAR}"
MY_FLOAT = "${shared.TEST_VAR_2}"
MY_ENV_SECRET = "${envs.MY_ENV_SECRET}"
MY_SHELL_VAR = "${`echo HELLO FROM SHELL`}"

[frontend.production]
NEXTAUTH_URL = "${shared.PROD_HOST}"
API_URL= "${shared.PROD_HOST}/api"
MONGODB_URI = "${shared.mongodb.production.MONGODB_URI}"

[backend]
JWT_SECRET = "${shared.JWT_SECRET}"
GOOGLE_ID = "${shared.GOOGLE_ID}"
GOOGLE_SECRET = "${shared.GOOGLE_SECRET}"

[backend.local]
# JWT_SECRET = "AMBARADILLA_LOCAL_JWT_SECRET"
BASE_URL = "http://localhost:8000/api"
FRONTEND_URL = "${frontend.local.NEXTAUTH_URL}"
MONGODB_URI = "${shared.mongodb.local.MONGODB_URI}"
PORTS = [
    "${shared.PORT}", 
    "4000"
]
MYSQL_ROOT_PASSWORD = "${shared.mysql.local.MYSQL_ROOT_PASSWORD}"
MYSQL_USER = "${shared.mysql.local.MYSQL_USER}"
MYSQL_DATABASE = "${shared.mysql.local.MYSQL_DATABASE}"
MYSQL_PASSWORD = "${shared.mysql.local.MYSQL_PASSWORD}"

[backend.production]
BASE_URL = "${shared.PROD_HOST}/api"
FRONTEND_URL = "${frontend.local.NEXTAUTH_URL}"
MONGODB_URI = "${shared.mongodb.local.MONGODB_URI}"
MYSQL_ROOT_PASSWORD = "${shared.mysql.prod.MYSQL_ROOT_PASSWORD}"
MYSQL_USER = "${shared.mysql.prod.MYSQL_USER}"
MYSQL_DATABASE = "${shared.mysql.prod.MYSQL_DATABASE}"
MYSQL_PASSWORD = "${shared.mysql.prod.MYSQL_PASSWORD}"
```

```bash
envsgen config.toml backend
```

**Output:**
```dotenv
GOOGLE_ID=google_oauth_client_id_example
GOOGLE_SECRET=google_oauth_client_secret_example
JWT_SECRET=demo_jwt_secret_change_me
```

**With `--expand`:**
```dotenv
GOOGLE_SECRET=google_oauth_client_secret_example
JWT_SECRET=demo_jwt_secret_change_me
LOCAL__MONGODB_URI=mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin
LOCAL__MYSQL_DATABASE=appdb
LOCAL__MYSQL_PASSWORD=appuserpass
LOCAL__MYSQL_ROOT_PASSWORD=rootpassword123
LOCAL__MYSQL_USER=appuser
LOCAL__PORTS=[8000 4000]
LOCAL__BASE_URL=http://localhost:8000/api
LOCAL__FRONTEND_URL=http://localhost:3000
PRODUCTION__MYSQL_PASSWORD=appuserprodpass
PRODUCTION__MYSQL_ROOT_PASSWORD=prodpassword123
PRODUCTION__MYSQL_USER=appuser
PRODUCTION__BASE_URL=https://www.mysite.com/api
PRODUCTION__FRONTEND_URL=http://localhost:3000
PRODUCTION__MONGODB_URI=mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin
PRODUCTION__MYSQL_DATABASE=appdb
GOOGLE_ID=google_oauth_client_id_example
```

**Inheritance of parent + overwrite values:**
```bash
envsgen config.toml backend.local
```

**Output:**
```dotenv
PORTS=[8000 4000]
FRONTEND_URL=http://localhost:3000
MONGODB_URI=mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin
GOOGLE_ID=google_oauth_client_id_example
MYSQL_USER=appuser
MYSQL_DATABASE=appdb
GOOGLE_SECRET=google_oauth_client_secret_example
MYSQL_PASSWORD=appuserpass
MYSQL_ROOT_PASSWORD=rootpassword123
JWT_SECRET=demo_jwt_secret_change_me
BASE_URL=http://localhost:8000/api
```

### 2. JSON

**Command:**
```bash
envsgen config.toml backend --json
```

**Output:**
```json
{
  "BASE_URL": "https://api.example.com/v1",
  "JWT_SECRET": "supersecret",
  "PORT": 8000
}
```

**With `--expand`:**
```json
{
  "GOOGLE_ID": "google_oauth_client_id_example",
  "GOOGLE_SECRET": "google_oauth_client_secret_example",
  "JWT_SECRET": "demo_jwt_secret_change_me",
  "local": {
    "BASE_URL": "http://localhost:8000/api",
    "FRONTEND_URL": "http://localhost:3000",
    "MONGODB_URI": "mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin",
    "MYSQL_DATABASE": "appdb",
    "MYSQL_PASSWORD": "appuserpass",
    "MYSQL_ROOT_PASSWORD": "rootpassword123",
    "MYSQL_USER": "appuser",
    "PORTS": [
      "8000",
      "4000"
    ]
  },
  "production": {
    "BASE_URL": "https://www.mysite.com/api",
    "FRONTEND_URL": "http://localhost:3000",
    "MONGODB_URI": "mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin",
    "MYSQL_DATABASE": "appdb",
    "MYSQL_PASSWORD": "appuserprodpass",
    "MYSQL_ROOT_PASSWORD": "prodpassword123",
    "MYSQL_USER": "appuser"
  }
}
```

### 3. YAML

**Command:**
```bash
envsgen config.toml backend --yaml
```

**Output:**
```yaml
BASE_URL: https://api.example.com/v1
JWT_SECRET: supersecret
PORT: 8000
```

**With `--expand`:**
```toml
GOOGLE_ID: google_oauth_client_id_example
GOOGLE_SECRET: google_oauth_client_secret_example
JWT_SECRET: demo_jwt_secret_change_me
local:
  BASE_URL: http://localhost:8000/api
  FRONTEND_URL: http://localhost:3000
  MONGODB_URI: mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin
  MYSQL_DATABASE: appdb
  MYSQL_PASSWORD: appuserpass
  MYSQL_ROOT_PASSWORD: rootpassword123
  MYSQL_USER: appuser
  PORTS:
    - "8000"
    - "4000"
production:
  BASE_URL: https://www.mysite.com/api
  FRONTEND_URL: http://localhost:3000
  MONGODB_URI: mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin
  MYSQL_DATABASE: appdb
  MYSQL_PASSWORD: appuserprodpass
  MYSQL_ROOT_PASSWORD: prodpassword123
  MYSQL_USER: appuser
```


### 4. Bash Export Script

**Command:**
```bash
envsgen config.toml backend --bash -o set-env.sh
```

**Output (set-env.sh):**
```bash
#!/bin/bash

export BASE_URL="https://api.example.com/v1"
export JWT_SECRET="supersecret"
export PORT="8000"
```

**With `--expand`:**
```bash
#!/bin/bash

export LOCAL__MONGODB_URI="mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin"
export LOCAL__MYSQL_DATABASE="appdb"
export LOCAL__MYSQL_PASSWORD="appuserpass"
export LOCAL__MYSQL_ROOT_PASSWORD="rootpassword123"
export LOCAL__MYSQL_USER="appuser"
export LOCAL__PORTS="[8000 4000]"
export LOCAL__BASE_URL="http://localhost:8000/api"
export LOCAL__FRONTEND_URL="http://localhost:3000"
export PRODUCTION__MYSQL_DATABASE="appdb"
export PRODUCTION__MYSQL_PASSWORD="appuserprodpass"
export PRODUCTION__MYSQL_ROOT_PASSWORD="prodpassword123"
export PRODUCTION__MYSQL_USER="appuser"
export PRODUCTION__BASE_URL="https://www.mysite.com/api"
export PRODUCTION__FRONTEND_URL="http://localhost:3000"
export PRODUCTION__MONGODB_URI="mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin"
export GOOGLE_ID="google_oauth_client_id_example"
export GOOGLE_SECRET="google_oauth_client_secret_example"
export JWT_SECRET="demo_jwt_secret_change_me"
```

**Usage:**
```bash
source set-env.sh
```

### 5. Docker Compose

**Command:**
```bash
envsgen config.toml docker.local --docker -o docker-compose.yml
```

See [Docker Compose Example](#docker-compose-generation) below.

### 6. Caddyfile

**Command:**
```bash
envsgen config.toml caddy --caddy -o Caddyfile
```

**Input TOML:**
```toml
[caddy."example.com"]
reverse_proxy = "localhost:8000"

[caddy."example.com".tls]
email = "admin@example.com"
```

**Output (Caddyfile):**
```
example.com {
	reverse_proxy localhost:8000
	tls {
		email admin@example.com
	}
}
```

#### Multiple Values For Same Key

##### The `_` (Underscore) Syntax for Caddyfiles

Caddy configuration often requires multiple directives with the same name. Since TOML only allows unique keys within a section, envsgen uses a special `_` key with an array:

```toml
# In your TOML config:
[caddy.prod."api.example.com"."reverse_proxy /api/*".to]
_ = [
    "http://backend1:8080",
    "http://backend2:8080",
    "http://backend3:8080",
]
```

This generates:

```caddyfile
api.example.com {
  reverse_proxy /api/* {
      to http://backend1:8080
      to http://backend2:8080
      to http://backend3:8080
  }
}
```

---

## 🔧 Advanced Features

### TOML Imports

Keep configurations modular by importing other TOML files:

**shared.toml:**
```toml
[shared]
JWT_SECRET = "supersecret"
PROD_HOST = "https://api.example.com"
```

**config.toml:**
```toml
#!import ./shared.toml
# or
#!import ./shared    # .toml extension auto-added

[backend]
BASE_URL = "${shared.PROD_HOST}/v1"
JWT_SECRET = "${shared.JWT_SECRET}"
```

**Features:**
- `.toml` extension is optional
- Relative paths resolved from config file directory
- Recursive imports supported (nested imports)
- Absolute paths supported

### Hierarchical Configuration

**Example structure:**
```toml
[app]
NAME = "MyApp"
VERSION = "1.0.0"

[app.database]
HOST = "localhost"
PORT = 5432

[app.database.credentials]
USER = "admin"
PASSWORD = "secret"
```

**With `--expand`:**
```bash
envsgen config.toml app --expand
```

**Output (dotenv):**
```dotenv
NAME=MyApp
VERSION=1.0.0
DATABASE__HOST=localhost
DATABASE__PORT=5432
DATABASE__CREDENTIALS__USER=admin
DATABASE__CREDENTIALS__PASSWORD=secret
```

---

## 💼 Real-World Examples

### Full-Stack Application Setup

**master.toml:**
```toml
[shared]
JWT_SECRET = "demo_jwt_secret_change_me"
PROD_HOST = "https://www.mysite.com"
GOOGLE_CLIENT_ID = "${envs.GOOGLE_OAUTH_ID}"
GOOGLE_CLIENT_SECRET = "${envs.GOOGLE_OAUTH_SECRET}"

[shared.database.local]
POSTGRES_USER = "devuser"
POSTGRES_PASSWORD = "devpass"
POSTGRES_DB = "myapp_dev"
DATABASE_URL = "postgresql://${shared.database.local.POSTGRES_USER}:${shared.database.local.POSTGRES_PASSWORD}@localhost:5432/${shared.database.local.POSTGRES_DB}"

[shared.database.prod]
POSTGRES_USER = "produser"
POSTGRES_PASSWORD = "${envs.DB_PASSWORD}"
POSTGRES_DB = "myapp_prod"
DATABASE_URL = "postgresql://${shared.database.prod.POSTGRES_USER}:${shared.database.prod.POSTGRES_PASSWORD}@db:5432/${shared.database.prod.POSTGRES_DB}"

[backend.local]
NODE_ENV = "development"
PORT = 8000
BASE_URL = "http://localhost:${backend.local.PORT}"
DATABASE_URL = "${shared.database.local.DATABASE_URL}"
JWT_SECRET = "${shared.JWT_SECRET}"
GOOGLE_CLIENT_ID = "${shared.GOOGLE_CLIENT_ID}"
GOOGLE_CLIENT_SECRET = "${shared.GOOGLE_CLIENT_SECRET}"

[backend.production]
NODE_ENV = "production"
PORT = 8000
BASE_URL = "${shared.PROD_HOST}"
DATABASE_URL = "${shared.database.prod.DATABASE_URL}"
JWT_SECRET = "${envs.JWT_SECRET_PROD}"
GOOGLE_CLIENT_ID = "${shared.GOOGLE_CLIENT_ID}"
GOOGLE_CLIENT_SECRET = "${shared.GOOGLE_CLIENT_SECRET}"

[frontend.local]
NEXT_PUBLIC_API_URL = "http://localhost:8000/api"
NEXT_PUBLIC_APP_URL = "http://localhost:3000"

[frontend.production]
NEXT_PUBLIC_API_URL = "${shared.PROD_HOST}/api"
NEXT_PUBLIC_APP_URL = "${shared.PROD_HOST}"

[docker.local.services.postgres]
image = "postgres:15-alpine"
container_name = "myapp_postgres"
restart = "unless-stopped"
ports = ["5432:5432"]
volumes = ["postgres_data:/var/lib/postgresql/data"]
networks = ["myapp"]

[docker.local.services.postgres.environment]
POSTGRES_USER = "${shared.database.local.POSTGRES_USER}"
POSTGRES_PASSWORD = "${shared.database.local.POSTGRES_PASSWORD}"
POSTGRES_DB = "${shared.database.local.POSTGRES_DB}"

[docker.local.volumes]
postgres_data = ""

[docker.local.networks.myapp]
driver = "bridge"
```

**Generate all configurations:**

```bash
# Set required environment variables
export GOOGLE_OAUTH_ID="your_google_client_id"
export GOOGLE_OAUTH_SECRET="your_google_client_secret"

# Backend configurations
envsgen master.toml backend.local -o backend/.env.local
envsgen master.toml backend.production -o backend/.env.production

# Frontend configurations
envsgen master.toml frontend.local -o frontend/.env.local
envsgen master.toml frontend.production -o frontend/.env.production

# Docker Compose
envsgen master.toml docker.local --docker -o docker-compose.local.yml

# Bash script for CI/CD
envsgen master.toml backend.production --bash -o scripts/set-prod-env.sh
```

### Docker Compose Generation

**Input (config.toml):**
```toml
[docker.local.services.caddy]
image = "caddy:latest"
restart = "unless-stopped"
cap_add = ["NET_ADMIN"]
ports = ["80:80", "443:443"]
volumes = ["./Caddyfile:/etc/caddy/Caddyfile", "caddy_data:/data"]
networks = ["web"]

[docker.local.services.app]
build = "./app"
depends_on = ["postgres"]
environment = { NODE_ENV = "development" }
ports = ["3000:3000"]
networks = ["web"]

[docker.local.services.postgres]
image = "postgres:15"
environment = { POSTGRES_PASSWORD = "devpass" }
volumes = ["postgres_data:/var/lib/postgresql/data"]
networks = ["web"]

[docker.local.volumes]
caddy_data = ""
postgres_data = ""

[docker.local.networks.web]
driver = "bridge"
```

**Generate Docker Compose:**
```bash
envsgen config.toml docker.local --docker -o docker-compose.yml
```

**Output (docker-compose.yml):**
```yaml
networks:
  web:
    driver: bridge
services:
  app:
    build: ./app
    depends_on:
      - postgres
    environment:
      NODE_ENV: development
    networks:
      - web
    ports:
      - 3000:3000
  caddy:
    cap_add:
      - NET_ADMIN
    image: caddy:latest
    networks:
      - web
    ports:
      - 80:80
      - 443:443
    restart: unless-stopped
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
  postgres:
    environment:
      POSTGRES_PASSWORD: devpass
    image: postgres:15
    networks:
      - web
    volumes:
      - postgres_data:/var/lib/postgresql/data
volumes:
  caddy_data: ""
  postgres_data: ""
```

### Multi-Environment Strategy

**Project structure:**
```
project/
├── configs/
│   ├── shared.toml
│   ├── secrets.local.toml
│   └── master.toml
├── backend/
│   ├── .env.local          # Generated
│   └── .env.production     # Generated
├── frontend/
│   ├── .env.local          # Generated
│   └── .env.production     # Generated
└── docker-compose.yml      # Generated
```

**Generation script (generate-configs.sh):**
```bash
#!/bin/bash

set -e

CONFIG="configs/master.toml"

echo "🔧 Generating configurations..."

# Backend
envsgen "$CONFIG" backend.local -o backend/.env.local
envsgen "$CONFIG" backend.staging -o backend/.env.staging
envsgen "$CONFIG" backend.production -o backend/.env.production

# Frontend
envsgen "$CONFIG" frontend.local -o frontend/.env.local
envsgen "$CONFIG" frontend.staging -o frontend/.env.staging
envsgen "$CONFIG" frontend.production -o frontend/.env.production

# Docker
envsgen "$CONFIG" docker.local --docker -o docker-compose.local.yml
envsgen "$CONFIG" docker.staging --docker -o docker-compose.staging.yml

echo "✅ All configurations generated successfully!"
```

---

## 🐛 Troubleshooting

### Common Issues

**1. Variable not found**
```
Error resolving variable 'shared.MISSING_KEY': key 'MISSING_KEY' not found
```
**Solution:** Ensure the key exists in your TOML file or use `--ignore-missing-vars`

**2. Section not found**
```
Error: section 'backend.typo' not found
```
**Solution:** Check section name spelling and TOML structure

**3. Shell command returns empty**
```
GIT_SHA=
```
**Solution:** Add `--allow-shell` flag to enable command execution

**4. Import file not found**
```
Error #!import file './missing.toml': no such file or directory
```
**Solution:** Check import path is correct relative to config file location

**5. Variable resolves to object**
```
Error: variable 'shared.database' resolves to an object, expected a primitive value
```
**Solution:** Reference specific keys within the object: `${shared.database.HOST}`

### Debug Tips

**Enable verbose mode:**
```bash
envsgen config.toml backend --verbose
```

**Test variable resolution:**
```bash
# Output JSON to see resolved values
envsgen config.toml backend --json --expand | jq
```

**Validate TOML syntax:**
```bash
# Use a TOML validator
toml-cli check config.toml
```

---

## 🛣️ Roadmap

### Features Ideas

- [ ] **Enhanced array handling**: CSV formatting and custom delimiters in dotenv output
- [ ] **Section name prefixes**: Optional inclusion of top-level section name in expanded keys
- [ ] **Explicit inheritance syntax**: Clearer inheritance control between sections
- [ ] **Conditional logic**: If/else expressions for environment-specific values
- [ ] **Loops and iteration**: Generate repeated configuration blocks
- [ ] **Template functions**: Built-in functions for common transformations (uppercase, lowercase, base64, etc.)
- [ ] **Validation rules**: Schema validation for configuration values
- [ ] **Encryption support**: Secure storage of sensitive values (similar to dotenvx)
- [ ] **Configuration merging**: Combine multiple TOML files with override priority
- [ ] **Watch mode**: Auto-regenerate outputs on configuration file changes

### Contributions Welcome

We welcome contributions! Areas where help is needed:
- Additional output format support (Kubernetes ConfigMap, Terraform tfvars, etc.)
- Improved documentation and examples
- Bug reports and feature requests
- Performance optimizations

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🤝 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/envsgen/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/envsgen/discussions)

---

**Made with ❤️ for developers who value DRY configuration management**