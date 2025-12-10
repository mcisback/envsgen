# Envsgen
Generate dotenv, JSON, or YAML from a single TOML configuration

Envsgen is a small CLI that reads a TOML file, resolves variables like ${path.to.value}, and emits one section as a flat environment map (dotenv) or as JSON/YAML. It also supports importing other TOML files, environment-variable lookups, and optional shell-command expansion.

---

## Features

- TOML → dotenv / JSON / YAML
- Select a specific TOML section (table) to output
- Variable interpolation inside strings with ${...}
  - Reference other keys by path: ${shared.JWT_SECRET}
  - Read environment variables: ${envs.MY_VAR}
  - Run shell commands: ${`echo hello`} (requires --allow-shell)
- Optional inclusion of child sections recursively (namespaced with double underscores) via --expand
- Simple import system: inline other TOML files with #!import path/to/file.toml
- Clean JSON (pretty-printed, no HTML escaping)
- YAML output (indented, clean)

Notes and caveats (by design, per current code):
- The “section” argument must refer to a TOML table (not a single key). It is required.
- In dotenv mode:
  - Keys are printed as-is by default. When using --expand, nested keys are uppercased and joined with double underscores (PARENT__CHILD__KEY).
  - Arrays print in Go’s default format (e.g., [a b c]); they are not comma-joined.
- String interpolation resolves one variable expression per value (e.g., "X=${A}/suffix" works; multiple different ${...} in the same string are not fully supported).

---

## Installation

Prerequisites:
- Go 1.18+ (1.20+ recommended)

Build locally:
- Clone your repository (or place the code in a folder), then:

```bash
go build -o envsgen .
```

Install to your PATH (two options):

- Using go install (requires module path set in go.mod):
```bash
go install ./
```

- Or copy the built binary to a directory on your PATH:
```bash
sudo cp envsgen /usr/local/bin
```

---

## Usage

```bash
envsgen <path/to/config.toml> <section> [options]
```

Options:
- --dotenv, -de      Output dotenv (default)
- --json, -j      Output JSON
- --yaml, -y      Output YAML
- --caddy, -cy				Output in CADDYFILE format (has some bugs but it works)
- --docker, -d				Output in DOCKER COMPOSE format
- --envs, -ev, --bash				Output a BASH script that sets env variables
- --output, -o PATH     Write to file instead of stdout
- --allow-shell     Allow execution of shell commands in ${`...`}
- --ignore-missing-vars, -iv      Ignore variables that do not resolve to anything
- --strict-vars-check, -sv			Stop if variables do not resolve to anything (default)
- --expand, -e      Include child sections recursively (namespaced keys)
- --verbose, -v			Be verbose

Example:
```bash
envsgen config.toml backend.local
envsgen config.toml backend --json
envsgen config.toml backend --yaml -o backend.yaml
envsgen config.toml backend --expand -o .env
envsgen config.toml backend --allow-shell
envsgen config.toml caddy.caddyfile --caddy
```

Notes:
- The section argument is required and must be a TOML table (e.g., backend, backend.local). A path that refers to a single primitive key (e.g., backend.local.BASE_URL) is not supported.
- The program exits with an error if the section does not exist, or if a variable path cannot be resolved or resolves to a table.

---

## TOML imports

You can inline other TOML files using a special directive:

```toml
#!import ./shared/shared
#!import ./secrets.toml
```

Details:
- The directive is: #!import <path>
- The .toml extension is optional in the directive; it will be added if missing.
- Relative paths are resolved relative to the main config file’s directory (or the current working directory if needed).
- Imports are processed recursively.

---

## Variable interpolation

You can interpolate variables inside strings using ${...}. Resolution rules:

- Reference other TOML keys by path (dot notation):
  - Example: BASE_URL = "${shared.PROD_HOST}/api"
- Read OS environment variables via the envs pseudo-root:
  - Example: SECRET = "${envs.MY_SUPER_SECRET}"
- Run shell commands (disabled by default; enable with --allow-shell):
  - Example: GIT_SHA = "${`git rev-parse --short HEAD`}"

Behavior:
- If a referenced key does not exist, the program exits with an error.
- If a reference resolves to an object (table), the program exits with an error. Only primitives are allowed.
- If --allow-shell is not provided, ${`...`} is replaced with an empty string and a warning is printed to stderr.
- Environment variables that are not set resolve to empty string.
- Interpolation is resolved recursively if the resolved value itself contains another ${...}.
- Each value supports one variable expression per string. Literal prefixes/suffixes work (e.g., "${A}/suffix"), but multiple different ${...} tokens in the same value are not fully supported.

---

## Selecting sections

The output is derived from exactly one selected section (a TOML table). Examples of valid sections:
- backend
- backend.local
- frontend.production

By default (without --expand), only the immediate keys inside that section are emitted (primitive values and arrays). Child tables are skipped.

If you pass --expand (or -e), child tables are included recursively. In dotenv mode, nested keys are uppercased and namespaced by double underscores. For example:
- Given a section backend that contains a child table local with key MONGODB_URI, the emitted key will be LOCAL__MONGODB_URI.
- Immediate keys in the selected section are uppercased too, but are not prefixed by the section name.
- Without expand if you print backend.local, backend.local will INHERIT parent backend section and overwrite parent section if same keys are present.

---

## Output formats

1) dotenv (default)
- Prints KEY=value lines.
- In default mode (no --expand), keys are printed as they appear in the TOML.
- With --expand, keys are uppercased and nested tables become namespaced via double underscores.
- Arrays print in Go’s default format (e.g., [a b c]).

2) JSON
- Pretty-printed with no HTML escaping.
- Values reflect the resolved data after interpolation.

3) YAML
- Indented, clean YAML.
- Values reflect the resolved data after interpolation.

---

## Examples

You can find more full example in the configs directory

Example config:
```toml
# configs/shared.toml
[shared]
PROD_HOST  = "https://mysite.com"
JWT_SECRET = "supersecret"
```

```toml
# configs/config.toml
#!import ./configs/shared

[backend]
BASE_URL   = "${shared.PROD_HOST}/api"
JWT_SECRET = "${shared.JWT_SECRET}"
FROM_ENV   = "${envs.MY_SUPER_SECRET}"
GIT_SHA    = "${`echo abc123`}" # requires --allow-shell

[backend.local]
MONGODB_URI = "mongodb://root:pass@localhost/mydb"
```

Set an env var:
```bash
export MY_SUPER_SECRET="aQCE3fJnIWajhIELeN5r2NGnWhZxhx"
```

- Dotenv for backend (immediate keys only):
```bash
envsgen config.toml backend
```
Produces:
```
BASE_URL=https://mysite.com/api
JWT_SECRET=supersecret
FROM_ENV=aQCE3fJnIWajhIELeN5r2NGnWhZxhx
GIT_SHA=
```
Note: GIT_SHA is empty without --allow-shell.

- Dotenv with shell enabled:
```bash
envsgen config.toml backend --allow-shell
```
Produces:
```
BASE_URL=https://mysite.com/api
JWT_SECRET=supersecret
FROM_ENV=aQCE3fJnIWajhIELeN5r2NGnWhZxhx
GIT_SHA=abc123
```

- Dotenv for backend with child sections included:
```bash
envsgen config.toml backend --expand
```
Produces:
```
BASE_URL=https://mysite.com/api
JWT_SECRET=supersecret
FROM_ENV=aQCE3fJnIWajhIELeN5r2NGnWhZxhx
GIT_SHA=
LOCAL__MONGODB_URI=mongodb://root:pass@localhost/mydb
```

- Dotenv for backend.local without --expands INHERITS parent:
backend.local vars WILL OVERWRITE backend VARS
```bash
envsgen config.toml backend.local
```
Produces:
```
## BACKEND
BASE_URL=https://mysite.com/api
JWT_SECRET=supersecret
FROM_ENV=aQCE3fJnIWajhIELeN5r2NGnWhZxhx
GIT_SHA=
## backend.local
MONGODB_URI=mongodb://root:pass@localhost/mydb
```

- JSON:
```bash
envsgen config.toml backend --json
```
Produces:
```json
{
  "BASE_URL": "https://mysite.com/api",
  "JWT_SECRET": "supersecret",
  "FROM_ENV": "aQCE3fJnIWajhIELeN5r2NGnWhZxhx",
  "GIT_SHA": ""
}
```

- YAML:
```bash
envsgen config.toml backend --yaml
```
Produces:
```yaml
BASE_URL: https://mysite.com/api
JWT_SECRET: supersecret
FROM_ENV: aQCE3fJnIWajhIELeN5r2NGnWhZxhx
GIT_SHA: ""
```

- Write to a file:
```bash
envsgen config.toml backend -o .env
envsgen config.toml backend.local -o .env.local
envsgen config.toml backend.production -o .env.production
```

## Docker Compose Generation

- You can use envsgen to generate docker-compose files, and share data between docker and backend, so you don't have to copy paste it 1000times in different files:

```toml
[shared]
JWT_SECRET = "demo_jwt_secret_change_me"
PROD_HOST = "https://www.mysite.com"
GOOGLE_ID = "google_oauth_client_id_example"
GOOGLE_SECRET = "google_oauth_client_secret_example"
RESEND_API_KEY = "resend_api_key_example"
TEST_VAR = 12312322
TEST_VAR_2 = 12.2012010
PORT = 8000

[shared.mysql.local]
MYSQL_ROOT_PASSWORD = "rootpassword123"
MYSQL_USER = "appuser"
MYSQL_PASSWORD = "appuserpass"
MYSQL_DATABASE = "appdb"

[shared.mysql.prod]
MYSQL_ROOT_PASSWORD = "prodpassword123"
MYSQL_USER = "appuser"
MYSQL_PASSWORD = "appuserprodpass"
MYSQL_DATABASE = "appdb"

[shared.mongodb.local]
MONGODB_URI = "mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin"

[shared.mongodb.production]
MONGODB_URI = "mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin"

[shared.stripe.test]
STRIPE_PUBLIC_KEY = "stripe_test_public_key_example"
STRIPE_SECRET_KEY = "stripe_test_secret_key_example"
STRIPE_WEBHOOK_SECRET = "stripe_test_webhook_secret_example"

[shared.stripe.live]
STRIPE_PUBLIC_KEY = "stripe_live_public_key_example"
STRIPE_SECRET_KEY = "stripe_live_secret_key_example"
STRIPE_WEBHOOK_SECRET = "stripe_live_webhook_secret_example"

[docker.local.services.caddy]
image = "caddy:latest"
restart = "unless-stopped"
cap_add = ["NET_ADMIN"]
ports = ["80:80", "443:443", "443:443/udp"]
volumes = [
  "./conf:/etc/caddy",
  # "$PWD/site:/srv",
  "caddy_data:/data",
  "caddy_config:/config"
]
networks = ["apps"]

[docker.local.services.mysql]
image = "mysql:8.0"
container_name = "mysql_main"
restart = "always"
ports = ["3306:3306"]
volumes = [ "mysql_data:/var/lib/mysql" ]
networks = ["apps"]

[docker.local.services.mysql.environment]
MYSQL_ROOT_PASSWORD = "${shared.mysql.local.MYSQL_ROOT_PASSWORD}"
MYSQL_USER = "${shared.mysql.local.MYSQL_USER}"
MYSQL_DATABASE = "${shared.mysql.local.MYSQL_DATABASE}"
MYSQL_PASSWORD = "${shared.mysql.local.MYSQL_PASSWORD}"

[docker.local.volumes]
caddy_data = ""
caddy_config = ""

[docker.local.networks.apps]
external = true

[docker.prod.services.mysql.environment]
MYSQL_ROOT_PASSWORD = "${shared.mysql.prod.MYSQL_ROOT_PASSWORD}"
MYSQL_USER = "${shared.mysql.prod.MYSQL_USER}"
MYSQL_DATABASE = "${shared.mysql.prod.MYSQL_DATABASE}"
MYSQL_PASSWORD = "${shared.mysql.prod.MYSQL_PASSWORD}"
```

```bash
envsgen config.toml docker.local --expand --yaml -o docker-compose.local.yaml
```
Produces a fully functional docker-compose file:
```yaml
networks:
  apps:
    external: true
services:
  caddy:
    cap_add:
      - NET_ADMIN
    image: caddy:latest
    networks:
      - apps
    ports:
      - 80:80
      - 443:443
      - 443:443/udp
    restart: unless-stopped
    volumes:
      - ./conf:/etc/caddy
      - caddy_data:/data
      - caddy_config:/config
  mysql:
    container_name: mysql_main
    environment:
      MYSQL_DATABASE: appdb
      MYSQL_PASSWORD: appuserpass
      MYSQL_ROOT_PASSWORD: rootpassword123
      MYSQL_USER: appuser
    image: mysql:8.0
    networks:
      - apps
    ports:
      - 3306:3306
    restart: always
    volumes:
      - mysql_data:/var/lib/mysql
volumes:
  caddy_config: ""
  caddy_data: ""
```

Now you can also create the dotenv for the backend:
```bash
envsgen config.toml backend.local
```

Produces:
```dotenv
GOOGLE_SECRET=google_oauth_client_secret_example
JWT_SECRET=demo_jwt_secret_change_me
MYSQL_DATABASE=appdb
BASE_URL=http://localhost:8000/api
MYSQL_USER=appuser
FRONTEND_URL=http://localhost:3000
MYSQL_PASSWORD=appuserpass
PORTS=[8000 4000]
MONGODB_URI=mongodb://USER:PASSWORD@localhost:27017/myAppDatabase?authSource=admin
MYSQL_ROOT_PASSWORD=rootpassword123
GOOGLE_ID=google_oauth_client_id_example
```

**So from one file you generated both docker-compose and backend dotenv file**

Now from the same file you could generate production and staging configuration

See the examples dir for more complete examples

---

## Errors and exit behavior

The program exits with a non-zero status on:
- Invalid TOML
- Missing section, or section path that does not refer to a table
- Variable path not found
- Variable resolves to a table (only primitives are allowed)
- File write errors

Warnings:
- Using ${`...`} without --allow-shell prints a warning to stderr and resolves to an empty string.

---

## Roadmap

- Improve array/object handling in dotenv output (e.g., CSV or configurable formatting)
- Optional inclusion of the top-level section name in dotenv keys when expanding
- Explicit inheritance
- If/Else conditions ?
- Loops ?

---

## License

MIT