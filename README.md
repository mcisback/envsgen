# Envsgen
## Generate multiple dotenv or configuration files from a single TOML file

Envsgen is a small command-line utility that converts a structured TOML configuration file into:
- .env (dotenv format)
- JSON
- YAML

It supports section selection, inheritance, and variable interpolation using the ${path.to.value} syntax, making it ideal for backend, frontend, Docker, CI/CD, and multi-environment setups.

---

## Features

- Convert TOML → env/json/yaml
  - Parses a TOML file, transforms it into JSON internally, and extracts flat environment maps.

- Section-based extraction
  - Target a specific section such as:
    - backend
    - backend.local
    - frontend.production
    - globals.stripe.test

- Variable expansion
  - Values may reference other fields:
    - ${globals.JWT_SECRET}
    - ${frontend.local.NEXTAUTH_URL}
  - Nested variables resolve recursively until a final primitive value is found.
  - Enviroments variable with ${envs.VAR_NAME}
  - Shell commands: with ${``}

- Supports primitive types
  - strings
  - booleans
  - integers / floats
  - timestamps
  - hex literals
  All are converted to strings safely.

- Outputs three formats
  - --dotenv (default)
  - --json
  - --yaml

---

## Installation

```bash
go install ./...
```

Or build manually:

```bash
go build -o envsgen .
sudo cp envsgen /usr/local/bin
```

---

## Usage

```bash
envsgen <path/to/config.toml> [section] [--dotenv|--json|--yaml]
```

---

## Examples

- Print full config as dotenv:
```bash
envsgen config.toml
```

- Print only frontend.local in dotenv format:
```bash
envsgen config.toml frontend.local
```

- Output to file
```bash
envsgen config.toml frontend.local -o .env
envsgen config.toml frontend.local -o .env.local
envsgen config.toml frontend.local -o whatever.txt
```

- Output JSON instead:
```bash
envsgen config.toml backend.local --json
```

- Output YAML:
```bash
envsgen config.toml backend --yaml
```

---

## TOML Structure Requirements

Your TOML file must include:
```toml
[globals]
```

Rules:
- The [globals] section must always be present.
- Referenced paths must resolve to primitive values, not tables.

---

## Variable Resolution

You can reference any value using ${path.to.key}:
```toml
HOST_URL = "${globals.PROD_HOST}"
DB_URI   = "${globals.mongodb.local.MONGODB_URI}"
```

You can reference environment variables with envs:
```bash
export MY_SUPER_SECRET="aQCE3fJnIWajhIELeN5r2NGnWhZxhx"
```

```toml
[globals]
PROD_HOST = "https://whatever"

[globals.mongodb.local]
MONGODB_URI = "mongodb://mongowhatever/myDatabase"

[backend]
HOST_URL = "${globals.PROD_HOST}"
DB_URI   = "${globals.mongodb.local.MONGODB_URI}"
MY_SUPER_SECRET = "${envs.MY_SUPER_SECRET}"
```

And you can run shell commands to get the output with the "--allow-shell" flag:
```toml
[globals]
PROD_HOST = "https://whatever"

[globals.mongodb.local]
MONGODB_URI = "mongodb://mongowhatever/myDatabase"

[backend]
HOST_URL = "${globals.PROD_HOST}"
DB_URI   = "${globals.mongodb.local.MONGODB_URI}"
MY_SHELL_COMMAND_OUTPUT = "${`echo Hello World`}"
MY_SHELL_COMMAND_SECRET = "${`/usr/local/bin/my-secret-manager-script get my-secret`}"
```

```bash
envsgen config.toml backend --allow-shell
```

Resolution rules:
- Errors if the referenced key does not exist
- Errors if the resolved value is an object (table)
- Resolves recursively (nested ${…} allowed)

---

## Output Formats

1) dotenv (default)
Produces:
```dotenv
KEY=value
OTHER_KEY=value2
```
- Arrays are serialized as comma-joined strings.

2) JSON
Pretty-printed, no HTML escaping:
```json
{
  "API_URL": "https://example.com",
  "DB_URI": "mongodb://..."
}
```

3) YAML
Indented, clean YAML:
```yaml
API_URL: https://example.com
DB_URI: mongodb://...
```

---

## Example Configuration

```toml
[globals]
JWT_SECRET = "supersecret"
PROD_HOST  = "https://mysite.com"

[backend]
BASE_URL   = "${globals.PROD_HOST}/api"
JWT_SECRET = "${globals.JWT_SECRET}"

[backend.local]
MONGODB_URI = "mongodb://root:pass@localhost/mydb"

[frontend.local]
NEXTAUTH_URL = "http://localhost:3000"
JWT_SECRET   = "${globals.JWT_SECRET}"
```

Running:
```bash
envsgen config.toml backend.local
```

Produces:
```dotenv
MONGODB_URI=mongodb://root:pass@localhost/mydb
```

---

## How It Works (Internally)

1. Load TOML → Go map
2. Convert TOML → JSON → back into map[string]any
3. Navigate to requested section
4. Flatten only primitive child values (section inheritance)
5. Resolve ${variables} recursively
6. Print in selected format

---

## ROADMAP
1. Better support to arrays, objects so you can generate docker-compose files

---

## License

MIT