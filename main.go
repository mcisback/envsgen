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

// TODO: add encryption like in here: https://dotenvx.com/docs/quickstart ?
/*
 * Basically with a strong password or public/private encryption it encrypts the dotenv
 * In this case the TOML
 * And the key is passed securely for example as a docker/os enviroment variable
 * Or use dotenvx directly
 */

var allowShell bool = false
var includeChildSections bool = false
var beVerbose = false
var ignoreMissingVars = false

func GetVariableValue(data any, path string) (any, error) {
	// Run when ${`shell command`} is used

	if strings.HasPrefix(path, "`") && strings.HasSuffix(path, "`") {
		parts := strings.Split(path, "`")
		parts = parts[1 : len(parts)-1]

		shellCmd := strings.Join(parts, "")

		if !allowShell {
			if beVerbose {
				fmt.Fprint(os.Stderr, "For command: ", shellCmd, "\nShell Command execution not allowed\nUse --allow-shell to enable it.\n\n")
			}

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

	return GetNodeFromPath(data, path)
}

func GetNodeFromPath(root any, path string) (any, error) {
	parts := strings.Split(path, ".")

	current := root
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

// func pathToVarValue(root any, path string) string {
// 	re := regexp.MustCompile(`\$\{([^}]+)\}`)

// 	path = re.ReplaceAllStringFunc(path, func(match string) string {

// 		fmt.Println("MATCH_VAR: ", match)

// 		matches := re.FindStringSubmatch(path)

// 		if len(matches) > 1 {
// 			varName := matches[1] // group 1 is the inside of ${...}

// 			varValue, err := GetVariableValue(root, varName)
// 			if err != nil {
// 				if ignoreMissingVars {

// 					if beVerbose {
// 						fmt.Fprintf(os.Stderr, "Error resolving variable '%s': %v\n", varName, err)
// 					}

// 					return matches[0]
// 				}

// 				fmt.Printf("Error resolving variable '%s': %v\n", varName, err)

// 				os.Exit(1)
// 			}

// 			switch varValue.(type) {
// 			case map[string]any:
// 				fmt.Printf("Error: variable '%s' resolves to an object, expected a primitive value\n", varName)
// 				os.Exit(1)
// 			case string:
// 				varValue = varValue.(string)

// 				// if the resolved string is again a variable, resolve it recursively
// 				if re.MatchString(varValue.(string)) {
// 					return pathToVarValue(root, varValue.(string))
// 				}

// 				return varValue.(string)
// 			case int, int64, uint64, uint16, uint32:
// 				return fmt.Sprintf("%d", varValue.(int))
// 			case float64, float32:
// 				varValue = fmt.Sprintf("%f", varValue.(float64))

// 				return ReplaceIfMatch(varValue.(string), `\.0+`, "")
// 			case bool:
// 				return fmt.Sprintf("%t", varValue.(bool))
// 				// allowed primitive types
// 			default:
// 				fmt.Printf("Error: variable '%s' is not supported\n", varName)
// 				os.Exit(1)
// 			}

// 			// fmt.Println("FOUND: ", matches[0], " => ", varName, " = ", varValue) // prints: var.name

// 			// value := strings.ReplaceAll(path, matches[0], fmt.Sprintf("%s", varValue))
// 		}

// 		return match
// 	})

// 	return path
// }

func pathToVarValue(root any, input string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	// Replace all occurrences like ${...} in the string
	resolved := re.ReplaceAllStringFunc(input, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}

		varName := sub[1]

		varValue, err := GetVariableValue(root, varName)
		if err != nil {
			if ignoreMissingVars {
				if beVerbose {
					fmt.Fprintf(os.Stderr, "Error resolving variable '%s': %v\n", varName, err)
				}
				// keep the original ${...} literal
				return match
			}

			fmt.Printf("Error resolving variable '%s': %v\n", varName, err)
			os.Exit(1)
		}

		switch v := varValue.(type) {
		case map[string]any:
			fmt.Printf("Error: variable '%s' resolves to an object, expected a primitive value\n", varName)
			os.Exit(1)

		case string:
			// recursively resolve nested variables inside the string
			return pathToVarValue(root, v)

		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			s := fmt.Sprint(v)
			// strip trailing ".0" / ".0000" etc
			return ReplaceIfMatch(s, `\.0+`, "")

		case bool:
			return fmt.Sprintf("%t", v)

		default:
			fmt.Printf("Error: variable '%s' is not supported (type %T)\n", varName, v)
			os.Exit(1)
		}

		return "" // unreachable, but satisfies the compiler
	})

	return resolved
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
			if includeChildSections {
				obj := parseVariables(root, v.(map[string]any)).(map[string]any)
				out[k] = obj
			} else {
				continue
			}

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
		case bool:
			value := v.(bool)
			out[k] = value
		case float32, float64:
			value := fmt.Sprintf("%f", v.(float64))
			out[k] = ReplaceIfMatch(value, `\.0+`, "")
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
	O_DOTENV OutputMode = "dotenv"
	O_JSON   OutputMode = "json"
	O_YAML   OutputMode = "yaml"
	O_BASH   OutputMode = "bash"
	O_CADDY  OutputMode = "caddy"
	O_DOCKER OutputMode = "docker"
)

func printUsageAndExit() {
	fmt.Printf("Usage:\n\t%s <path/to/config.toml> [section] [options]\n\n", filepath.Base(os.Args[0]))
	fmt.Println("Options:")
	fmt.Println("\t--json, -j				Output in JSON format")
	fmt.Println("\t--dotenv, -de				Output in DOTENV format (default)")
	fmt.Println("\t--yaml, -y				Output in YAML format")
	fmt.Println("\t--caddy, -cy				Output in CADDYFILE format (beta)")
	fmt.Println("\t--docker, -d				Output in DOCKER COMPOSE format")
	fmt.Println("\t--envs, -ev, --bash				Output a BASH script that sets env variables")
	fmt.Println("\t--allow-shell			Allow execution of shell commands")
	fmt.Println("\t--ignore-missing-vars, -iv			Ignore variables that do not resolve to anything")
	fmt.Println("\t--strict-vars-check, -sv			Stop if variables do not resolve to anything (default)")
	fmt.Println("\t--output, -o filepath			Output to file instead of stdout")
	fmt.Println("\t--verbose, -v			Be verbose")

	os.Exit(0)
}

func main() {
	var section string

	outputMode := O_DOTENV // default mode
	outputFile := os.Stdout

	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	configFilePath := os.Args[1]
	if len(os.Args) >= 3 {
		section = os.Args[2]
	}

	if section == "" {
		printUsageAndExit()
	}

	if strings.HasPrefix(section, "-") {
		fmt.Fprintf(os.Stderr, "section %s is probably a flag option and not a proper section, exiting...\n", section)

		os.Exit(1)
	}

	args := os.Args[3:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--help", "-h":
			printUsageAndExit()

		case "--verbose", "-v":
			beVerbose = true

		case "--json", "-j":
			outputMode = O_JSON

		case "--dotenv", "-de":
			outputMode = O_DOTENV

		case "--yaml", "-y":
			outputMode = O_YAML

		case "--caddy", "-cy":
			outputMode = O_CADDY
			includeChildSections = true

		case "--docker", "-d":
			outputMode = O_DOCKER
			includeChildSections = true
			ignoreMissingVars = true // your note

		case "--envs", "-ev", "--bash":
			outputMode = O_BASH

		case "--output", "-o":
			if i+1 >= len(args) {
				fmt.Printf("Error: %s requires a file path\n", arg)
				os.Exit(1)
			}

			outPath := args[i+1]

			if strings.HasPrefix(outPath, "-") {
				fmt.Printf("Error: %s requires a file path, but got flag '%s'\n", arg, outPath)
				os.Exit(1)
			}

			file, err := os.Create(outPath)
			if err != nil {
				fmt.Printf("Error: cannot write to file '%s': %v\n", outPath, err)
				os.Exit(1)
			}

			defer file.Close()
			outputFile = file

			i++ // skip filename
			continue

		case "--allow-shell":
			allowShell = true

		case "--ignore-missing-vars", "-iv":
			ignoreMissingVars = true

		case "--strict-vars-check", "-sv":
			ignoreMissingVars = false

		case "--expand", "-e":
			includeChildSections = true
		}

	}

	raw, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", configFilePath, err)
		os.Exit(1)
	}

	raw = []byte(preProcessTOML(configFilePath, string(raw)))

	// fmt.Println(string(raw))

	// return

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

	// 4) If a section is provided, drill into it using JSON structure
	var toPrint any = root

	// FIX: go run . masterenvs.toml backend.local.BASE_URL prints but should not

	finalObj := make(map[string]any)

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

		for k, v := range next.(map[string]any) {
			switch v.(type) {
			case map[string]any:
				// skip child sections
				if includeChildSections {
					// wildcard section, include child sections
					finalObj[k] = v
				} else {
					continue
				}
			// TODO: support direct variable resolution ?
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

	if includeChildSections {
		toPrint = parseVariables(root, current)
	} else {
		toPrint = parseVariables(root, finalObj)
	}

	switch outputMode {
	case O_JSON:
		// print JSON
		printJSON(toPrint, outputFile)
	case O_DOTENV:
		// print dotenv
		printDotEnv("", toPrint, outputFile)
	case O_YAML:
		printYAML(toPrint, outputFile)
	case O_BASH:
		printBASH("", toPrint, outputFile)
	case O_CADDY:
		printCADDY(toPrint, outputFile)
	case O_DOCKER:
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

func printDotEnv(prefix string, data any, outputFile io.Writer) {
	obj, ok := data.(map[string]any)
	if !ok {
		fmt.Fprintln(outputFile, "Error in printDotEnv: data is not an object")
		os.Exit(1)
	}

	for k, v := range obj {
		if includeChildSections {
			pre := strings.ToUpper(k)

			if prefix != "" {
				pre = prefix + "__" + strings.ToUpper(k)
			}
			// skip child sections
			switch v.(type) {
			case map[string]any:
				printDotEnv(pre, v, outputFile)
			default:
				fmt.Fprintf(outputFile, "%s=%v\n", pre, v)
			}

			continue
		}

		fmt.Fprintf(outputFile, "%s=%v\n", prefix+k, v)
	}
}

func printCADDY(data any, outputFile io.Writer) {
	obj, ok := data.(map[string]any)
	if !ok {
		fmt.Fprintln(outputFile, "Error in printCADDY: data is not an object")
		os.Exit(1)
	}

	for domain, v := range obj {
		switch v.(type) {
		case map[string]any:

			printCADDYSection(domain, v.(map[string]any), outputFile)

		default:
			fmt.Fprintf(os.Stderr, "%s %v is not an object\n", domain, v)

			os.Exit(1)
		}
	}
}

//	domain {
//		directives
//	}
func printCADDYSection(domain string, parent map[string]any, outputFile io.Writer) {
	fmt.Fprintf(outputFile, "%s {\n", domain)

	for sectionName, sectionContent := range parent {
		switch sectionContent.(type) {
		case map[string]any:

			printCADDYBlock("\t", sectionName, sectionContent.(map[string]any), outputFile)

		case []any:
			printCADDYArray("\t", sectionName, sectionContent.([]any), outputFile)
		default:
			if sectionContent == "" {
				fmt.Fprintf(outputFile, "\t%s\n", sectionName)

				break
			}
			fmt.Fprintf(outputFile, "\t%s %v\n", sectionName, sectionContent)
		}
	}

	fmt.Fprintf(outputFile, "}\n\n")
}

func printCADDYBlock(level string, blockName string, block map[string]any, outputFile io.Writer) {

	if raw, ok := block["_"]; ok {
		// key exists
		switch v := raw.(type) {
		case []any:
			fmt.Fprintln(outputFile)

			// handle the special "_" array
			for _, item := range v {
				fmt.Fprintf(outputFile, "%s%s %v\n", level, blockName, item)
			}

			fmt.Fprintln(outputFile)
		default:
			// handle unexpected type
			fmt.Fprintf(outputFile, "%s\t# unexpected '_' type: %T\n", level, raw)
		}

		return
	}

	fmt.Fprintf(outputFile, "%s%s {\n", level, blockName)

	for directive, value := range block {
		switch value.(type) {
		case map[string]any:
			printCADDYBlock(level+"\t", directive, value.(map[string]any), outputFile)
		case []any:

			printCADDYArray(level+"\t", directive, value.([]any), outputFile)
		default:
			if value == "" {
				fmt.Fprintf(outputFile, "%s%s\n", level+"\t", directive)

				break
			}
			fmt.Fprintf(outputFile, "%s%s %v\n", level+"\t", directive, value)
		}
	}

	fmt.Fprintf(outputFile, "%s}\n", level)
}

func printCADDYArray(level string, directive string, array []any, outputFile io.Writer) {
	fmt.Fprintf(outputFile, "%s%s", level, directive)

	for _, v := range array {
		fmt.Fprintf(outputFile, " %v", v)
	}

	fmt.Fprintln(outputFile)
}

func printBASH(prefix string, data any, outputFile io.Writer) {
	obj, ok := data.(map[string]any)
	if !ok {
		fmt.Fprintln(outputFile, "Error in printBASH: data is not an object")
		os.Exit(1)
	}

	// First run
	if prefix == "" {
		fmt.Fprintf(outputFile, "#!/bin/bash\n\n")
	}

	for k, v := range obj {
		if includeChildSections {
			pre := strings.ToUpper(k)

			if prefix != "" {
				pre = prefix + "__" + strings.ToUpper(k)
			}
			// skip child sections
			switch v.(type) {
			case map[string]any:
				printBASH(pre, v, outputFile)
			default:
				fmt.Fprintf(outputFile, "export %s=\"%v\"\n", pre, v)
			}

			continue
		}

		fmt.Fprintf(outputFile, "export %s=\"%v\"\n", prefix+k, v)
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

func preProcessTOML(configFilePath string, raw string) string {
	re := regexp.MustCompile(`#!import\s+(.+)`)

	// Find all matches; each entry contains full match + captured group(s)
	matches := re.FindAllStringSubmatch(raw, -1)

	for _, m := range matches {
		// m[0] is the entire "#!import file"
		// m[1] is the extracted "file"
		fileToImport := strings.TrimSpace(m[1]) // avoid import file' ' <- sneaky space

		if !strings.HasSuffix(fileToImport, ".toml") {
			fileToImport = fileToImport + ".toml"
		}

		dirname := filepath.Dir(configFilePath)

		// Not an absolute path
		if !strings.HasPrefix(fileToImport, "/") {

			if !strings.HasPrefix(dirname, "/") {
				wd, _ := os.Getwd()

				dirname = wd + "/" + dirname
			}

			fileToImport = filepath.Join(dirname, fileToImport)
		}

		// fmt.Println(fileToImport)

		// os.Exit(1)

		toml, err := os.ReadFile(fileToImport)
		if err != nil {
			fmt.Printf("Error #!import file '%s': %v\n", fileToImport, err)
			os.Exit(1)
		}

		// Recurse if imports nest like curious little Matryoshkas
		if re.Match(toml) {
			toml = []byte(preProcessTOML(configFilePath, string(toml)))
		}

		// Replace this specific import statement with file contents
		raw = strings.ReplaceAll(raw, m[0], string(toml)+"\n")
	}

	return raw
}
