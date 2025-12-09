package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var allowShell bool = false

func GetVariableValue(data any, path string) (any, error) {
	// Run when ${`shell command`} is used

	if strings.HasPrefix(path, "`") && strings.HasSuffix(path, "`") {
		parts := strings.Split(path, "`")
		parts = parts[1 : len(parts)-1]

		shellCmd := strings.Join(parts, "")

		if !allowShell {
			fmt.Fprint(os.Stderr, "For command: ", shellCmd, "\nShell Command execution not allowed\nUse --allow-shell to enable it.\n\n")

			return "", nil
		}

		cmd := exec.Command("bash", "-c", shellCmd)
		out, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Println("Error:", err)

			os.Exit(1)
		}

		// NOTE: maybe remove just newlines
		return strings.TrimSpace(string(out)), nil
	}

	parts := strings.Split(path, ".")

	if parts[0] == "envs" {
		parts = parts[1:]
		envVar := strings.Join(parts, "")

		return os.Getenv(envVar), nil
	}

	// if parts[0] == "shell" {
	// 	parts = parts[1:]
	// 	shellCmd := strings.Join(parts, "")

	// 	cmd := exec.Command("bash", "-c", shellCmd)
	// 	out, err := cmd.CombinedOutput()

	// 	if err != nil {
	// 		fmt.Println("Error:", err)

	// 		os.Exit(1)
	// 	}

	// 	// NOTE: maybe remove just newlines
	// 	return strings.TrimSpace(string(out)), nil
	// }

	current := data
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("path '%s' is not an object", part)
		}

		v, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("key '%s' not found", part)
		}

		current = v
	}

	return current, nil
}

func ReplaceIfMatch(s, pattern, replaceWith string) string {
	re := regexp.MustCompile(pattern)

	if !re.MatchString(s) {
		return s
	}

	out := re.ReplaceAllString(s, replaceWith)
	return out
}

func pathToVarValue(root any, path string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	matches := re.FindStringSubmatch(path)

	if len(matches) > 1 {
		varName := matches[1] // group 1 is the inside of ${...}

		varValue, err := GetVariableValue(root, varName)
		if err != nil {
			fmt.Printf("Error resolving variable '%s': %v\n", varName, err)

			os.Exit(1)
		}

		switch varValue.(type) {
		case map[string]any:
			fmt.Printf("Error: variable '%s' resolves to an object, expected a primitive value\n", varName)
			os.Exit(1)
		case string:
			varValue = varValue.(string)

			// if the resolved string is again a variable, resolve it recursively
			if re.MatchString(varValue.(string)) {
				varValue = pathToVarValue(root, varValue.(string))
			}
		case int, int64, uint64, uint16, uint32:
			varValue = fmt.Sprintf("%d", varValue.(int))
		case float64, float32:
			varValue = fmt.Sprintf("%f", varValue.(float64))
			varValue = ReplaceIfMatch(varValue.(string), `\.0+`, "")
		case bool:
			varValue = fmt.Sprintf("%t", varValue.(bool))
			// allowed primitive types
		default:
			fmt.Printf("Error: variable '%s' is not supported\n", varName)
			os.Exit(1)
		}

		// fmt.Println("FOUND: ", matches[0], " => ", varName, " = ", varValue) // prints: var.name

		value := strings.ReplaceAll(path, matches[0], fmt.Sprintf("%s", varValue))

		return value
	}

	return path
}

func parseVariables(root any, node any) any {
	m, ok := node.(map[string]any)
	if !ok {
		return node
	}

	out := make(map[string]any)

	for k, v := range m {
		// keep only primitive values
		switch v.(type) {
		case map[string]any:
			// fmt.Println("IS MAP ", k)
			// skip child sections
			continue
		case []any:
			// fmt.Println("IS ARRAY ", k)
			arr := v.([]any)
			newArr := make([]any, len(arr))

			for i, item := range arr {
				switch item.(type) {
				case string:
					value := pathToVarValue(root, item.(string))
					newArr[i] = value
				default:
					newArr[i] = item
				}
			}

			out[k] = newArr
		default:
			value := v.(string)

			value = pathToVarValue(root, value)

			out[k] = value
		}
	}
	return out
}

type OutputMode string

const (
	O_DotEnv OutputMode = "dotenv"
	O_JSON   OutputMode = "json"
	O_YAML   OutputMode = "yaml"
)

func printUsage() {
	fmt.Printf("Usage:\n\t%s <path/to/config.toml> [section] [options]\n\n", filepath.Base(os.Args[0]))
	fmt.Println("Options:")
	fmt.Println("\t--json, -j				Output in JSON format")
	fmt.Println("\t--dotenv, -d				Output in dotenv format (default)")
	fmt.Println("\t--yaml, -y				Output in YAML format")
	fmt.Println("\t--allow-shell			Allow execution of shell commands")
	fmt.Println("\t--output, -o filepath			Output to file instead of stdout")

	os.Exit(0)
}

func main() {
	outputMode := O_DotEnv // default mode
	outputFile := os.Stdout

	if len(os.Args) < 2 {
		printUsage()
	}

	path := os.Args[1]
	var section string
	if len(os.Args) >= 3 {
		section = os.Args[2]
	}

	for i, arg := range os.Args {
		if arg == "--help" || arg == "-h" {
			printUsage()
		}

		if arg == "--json" || arg == "-j" {
			outputMode = O_JSON
		}

		if arg == "--dotenv" || arg == "-d" {
			outputMode = O_DotEnv
		}

		if arg == "--yaml" || arg == "-y" {
			outputMode = O_YAML
		}

		if arg == "--output" || arg == "-o" {
			if i+1 >= len(os.Args) {
				fmt.Printf("Error: %s requires a file path\n", arg)
				os.Exit(1)
			}

			outPath := os.Args[i+1]

			file, err := os.Create(outPath)
			if err != nil {
				fmt.Printf("Error: cannot write to file '%s': %v\n", outPath, err)
				os.Exit(1)
			}

			defer file.Close()
			outputFile = file
		}

		if arg == "--allow-shell" {
			allowShell = true
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", path, err)
		os.Exit(1)
	}

	// 1) TOML -> Go structure
	var tomlData map[string]any
	if err := toml.Unmarshal(raw, &tomlData); err != nil {
		fmt.Printf("Error parsing TOML: %v\n", err)
		os.Exit(1)
	}

	// 2) Go (TOML) -> JSON bytes
	jsonBytes, err := json.Marshal(tomlData)
	if err != nil {
		fmt.Printf("Error converting TOML to JSON: %v\n", err)
		os.Exit(1)
	}

	// 3) JSON bytes -> generic JSON data (interface{})
	var data any
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		os.Exit(1)
	}

	// Root must be an object
	root, ok := data.(map[string]any)
	if !ok {
		fmt.Println("Error: root of JSON is not an object")
		os.Exit(1)
	}

	// Validate [globals] exists in JSON
	globals, ok := root["globals"]
	if !ok {
		fmt.Println("Error: Missing [globals] section in the TOML/JSON.")
		os.Exit(1)
	}
	if _, ok := globals.(map[string]any); !ok {
		fmt.Println("Error: [globals] section must be an object.")
		os.Exit(1)
	}

	// 4) If a section is provided, drill into it using JSON structure
	var toPrint any = root

	// FIX: go run . masterenvs.toml backend.local.BASE_URL prints but should not

	finalObj := make(map[string]any)

	if section == "" {
		toPrint = root
		finalObj = root
	} else {
		parts := strings.Split(section, ".")

		current := any(root)
		for _, part := range parts {
			obj, ok := current.(map[string]any)
			if !ok {
				fmt.Printf("Error: section '%s' is not an object\n", section)
				os.Exit(1)
			}

			next, exists := obj[part]
			if !exists {
				fmt.Printf("Error: section '%s' not found\n", section)
				os.Exit(1)
			}

			// fmt.Println("HERE obj: ", part)

			// for k, v := range next.(map[string]any) {
			// 	fmt.Printf("key=%s type=%T\n", k, v)
			// }

			for k, v := range next.(map[string]any) {
				switch v.(type) {
				case map[string]any:
					// fmt.Println("Skipping ", k, "from output")
					// skip child sections
					continue
				// TODO: support direct variable resolution
				// case string:
				// 	value := pathToVarValue(root, v.(string))
				// 	fmt.Println(value)

				// 	return
				case []any:
					// fmt.Println("Array ", k)
					finalObj[k] = v
				default:
					finalObj[k] = v
					// fmt.Println("finalObj[", k, "] = ", v) // <- prints nothing
				}
			}

			current = next
		}

		// toPrint = stripChildSections(toPrint)
	}

	toPrint = parseVariables(root, finalObj)

	switch outputMode {
	case O_JSON:
		// print JSON
		printJSON(toPrint, outputFile)
	case O_DotEnv:
		// print dotenv
		printDotEnv(toPrint, outputFile)
	case O_YAML:
		// print dotenv
		printYAML(toPrint, outputFile)
	default:
		fmt.Printf("Error: unknown output mode '%s'\n", outputMode)
		os.Exit(1)
	}
}

func printJSON(data any, outputFile io.Writer) {
	encoder := json.NewEncoder(outputFile)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		fmt.Printf("Error printing JSON: %v\n", err)
		os.Exit(1)
	}
}

func printDotEnv(data any, outputFile io.Writer) {
	obj, ok := data.(map[string]any)
	if !ok {
		fmt.Fprintln(outputFile, "Error in printDotEnv: data is not an object")
		os.Exit(1)
	}

	for k, v := range obj {
		fmt.Fprintf(outputFile, "%s=%v\n", k, v)
	}
}

func printYAML(data any, outputFile io.Writer) {
	encoder := yaml.NewEncoder(outputFile)
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		fmt.Printf("Error printing YAML: %v\n", err)
		os.Exit(1)
	}
}
