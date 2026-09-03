package main

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"os"
	"path/filepath"
	"strings"

	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/connection"
	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/documents"

	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types"
	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types/schemas"
	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v3"
)

var (
	client       *types.Client
	expandedMode bool
	vaultHeader  = "$AMARA_VAULT;1.1;AES256\n"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	cfg, err := loadCLICredentials()
	if err != nil {
		slog.Error("Could not load credentials cleanly", "error", err)
		os.Exit(1)
	}

	c, err := connection.NewClient(cfg)
	if err != nil {
		slog.Error("Client initialization failed", "error", err)
		os.Exit(1)
	}
	client = c

	fmt.Println("Amara Shell (aql) v1.0.0")
	fmt.Println("Type 'help' for available commands, '\\x' for expanded display, '\\q' or 'exit' to quit.")

	defer func() {
		if r := recover(); r != nil {
			cleanExit()
		}
	}()

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("aql=> "),
		prompt.OptionTitle("aql"),
	)
	p.Run()
}

func loadCLICredentials() (types.Config, error) {
	secKeyBase64 := os.Getenv("AMARA_SECURITY_KEY")
	if secKeyBase64 == "" {
		slog.Error("AMARA_SECURITY_KEY is missing. Boot aborted")
		os.Exit(1)
	}

	if len(secKeyBase64) != 64 {
		// #nosec G706
		slog.Error(fmt.Sprintf("AMARA_SECURITY_KEY has invalid length of %v. Must be a 64-character base64 string. Boot aborted", len(secKeyBase64)))
		os.Exit(1)
	}

	secKey, err := base64.StdEncoding.DecodeString(secKeyBase64)
	if err != nil {
		return types.Config{}, fmt.Errorf("invalid base64 in AMARA_SECURITY_KEY: %w", err)
	}

	if len(secKey) > 32 {
		secKey = secKey[:32]
	}

	cfg := types.Config{}

	apiEnv := os.Getenv("AMARA_TYPESENSE_API_KEY")
	nodesEnv := os.Getenv("AMARA_TYPESENSE_NODES")
	if apiEnv != "" && nodesEnv != "" {
		cfg.APIKey = apiEnv
		cfg.Nodes = strings.Split(nodesEnv, ",")
		return cfg, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, err
	}
	dirPath := filepath.Join(home, ".amara")
	path := filepath.Join(dirPath, "creds.json")

	fileInfo, err := os.Stat(path)
	if err == nil && !fileInfo.IsDir() {
		// #nosec G304
		encryptedBytes, err := os.ReadFile(path)
		if err != nil {
			return cfg, err
		}

		decryptedData, err := decryptAnsibleVaultStyle(string(encryptedBytes), secKey)
		if err != nil {
			return cfg, fmt.Errorf("failed to decrypt creds.json: %w", err)
		}

		if err := json.Unmarshal(decryptedData, &cfg); err != nil {
			return types.Config{}, fmt.Errorf("invalid JSON in decrypted creds.json: %w", err)
		}
		return cfg, nil
	}

	fmt.Println("Credentials not found. Entering configuration edit mode...")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Typesense API Key: ")
	apiKey, _ := reader.ReadString('\n')
	cfg.APIKey = strings.TrimSpace(apiKey)

	fmt.Print("Enter Typesense Nodes (comma-separated, e.g., http://node1:8108,http://node2:8108): ")
	nodesStr, _ := reader.ReadString('\n')
	nodes := strings.Split(strings.TrimSpace(nodesStr), ",")
	for i := range nodes {
		nodes[i] = strings.TrimSpace(nodes[i])
	}
	cfg.Nodes = nodes

	// #nosec G117
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return cfg, fmt.Errorf("failed to marshal config: %w", err)
	}

	encryptedVaultData, err := encryptAnsibleVaultStyle(cfgBytes, secKey)
	if err != nil {
		return cfg, fmt.Errorf("encryption failed: %w", err)
	}

	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return cfg, err
	}

	// #nosec G703
	if err := os.WriteFile(path, []byte(encryptedVaultData), 0600); err != nil {
		return cfg, err
	}

	fmt.Println("Credentials securely saved to ~/.amara/creds.json")
	return cfg, nil
}

func encryptAnsibleVaultStyle(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	hexData := hex.EncodeToString(ciphertext)

	var formattedHex strings.Builder
	for i := 0; i < len(hexData); i += 80 {
		end := min(i+80, len(hexData))
		formattedHex.WriteString(hexData[i:end])
		formattedHex.WriteString("\n")
	}

	return vaultHeader + formattedHex.String(), nil
}

func decryptAnsibleVaultStyle(vaultData string, key []byte) ([]byte, error) {
	if !strings.HasPrefix(vaultData, vaultHeader) {
		return nil, fmt.Errorf("missing or invalid AMARA_VAULT header")
	}

	hexData := strings.TrimPrefix(vaultData, vaultHeader)
	hexData = strings.ReplaceAll(hexData, "\n", "")
	hexData = strings.ReplaceAll(hexData, "\r", "")

	ciphertext, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	switch in {
	case "exit", "quit", "\\q":
		cleanExit()
	case "\\x":
		expandedMode = !expandedMode
		if expandedMode {
			fmt.Println("Expanded display is on.")
		} else {
			fmt.Println("Expanded display is off.")
		}
		return
	case "\\config":
		secKeyBase64 := os.Getenv("AMARA_SECURITY_KEY")
		if secKeyBase64 == "" || len(secKeyBase64) != 64 {
			fmt.Println("Error: AMARA_SECURITY_KEY is missing or invalid. Cannot edit configuration.")
			return
		}
		secKey, _ := base64.StdEncoding.DecodeString(secKeyBase64)
		if len(secKey) > 32 {
			secKey = secKey[:32]
		}

		home, err := os.UserHomeDir()
		if err == nil {
			dirPath := filepath.Join(home, ".amara")
			path := filepath.Join(dirPath, "creds.json")

			newCfg, err := runInteractiveSetup(secKey, dirPath, path)
			if err != nil {
				fmt.Printf("Failed to save configuration: %v\n", err)
				return
			}

			// Hot-swap the client instance so the new credentials take effect immediately
			newClient, err := connection.NewClient(newCfg)
			if err != nil {
				fmt.Printf("Warning: Failed to initialize new client connection: %v\n", err)
			} else {
				client = newClient
				fmt.Println("Client connection successfully re-established with new credentials.")
			}
		}
		return
	}

	args := append([]string{"aql"}, parseArgs(in)...)
	cmd := buildCLI()

	if err := cmd.Run(context.Background(), args); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runInteractiveSetup prompts the user for configuration, validates connectivity,
// encrypts it, saves it to disk, and returns the newly minted config.
func runInteractiveSetup(secKey []byte, dirPath, filePath string) (types.Config, error) {
	fmt.Println("Entering configuration edit mode...")
	reader := bufio.NewReader(os.Stdin)
	cfg := types.Config{}

	// Loop until the user provides valid, reachable credentials
	for {
		fmt.Print("Enter Typesense API Key: ")
		apiKey, _ := reader.ReadString('\n')
		cfg.APIKey = strings.TrimSpace(apiKey)

		fmt.Print("Enter Typesense Nodes (comma-separated, e.g., http://node1:8108,http://node2:8108): ")
		nodesStr, _ := reader.ReadString('\n')
		nodes := strings.Split(strings.TrimSpace(nodesStr), ",")

		var cleanedNodes []string
		for _, n := range nodes {
			cleaned := strings.TrimSpace(n)
			if cleaned != "" {
				cleanedNodes = append(cleanedNodes, cleaned)
			}
		}
		cfg.Nodes = cleanedNodes

		if len(cfg.Nodes) == 0 {
			fmt.Println("Error: At least one node URL is required. Please try again.")
			continue
		}

		fmt.Println("Testing connectivity to provided nodes...")
		if err := validateNodes(cfg.Nodes, cfg.APIKey); err != nil {
			fmt.Printf("Connectivity test failed: %v\n", err)
			fmt.Println("Please verify your API key and Node URLs, ensuring the Typesense cluster is running, and try again.")
			continue
		}

		fmt.Println("All nodes successfully validated!")
		break // Exit the loop since validation passed
	}

	// #nosec G117
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return cfg, fmt.Errorf("failed to marshal config: %w", err)
	}

	encryptedVaultData, err := encryptAnsibleVaultStyle(cfgBytes, secKey)
	if err != nil {
		return cfg, fmt.Errorf("encryption failed: %w", err)
	}

	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return cfg, err
	}

	// #nosec G703
	if err := os.WriteFile(filePath, []byte(encryptedVaultData), 0600); err != nil {
		return cfg, err
	}

	fmt.Println("Credentials securely saved to", filePath)
	return cfg, nil
}

func buildCLI() *cli.Command {
	return &cli.Command{
		Name:  "aql",
		Usage: "Typesense interactive CLI",
		Commands: []*cli.Command{
			{
				Name:  "collections",
				Usage: "Manage Typesense collections: collections [list|get <name>|delete <name>]",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sub := cmd.Args().First()
					if sub == "" || sub == "list" {
						res, err := client.Collections.RetrieveAll(ctx)
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					if sub == "get" {
						name := cmd.Args().Get(1)
						res, err := client.Collections.Retrieve(ctx, name)
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					if sub == "delete" {
						name := cmd.Args().Get(1)
						res, err := client.Collections.Delete(ctx, name)
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					return fmt.Errorf("unknown subcommand: %s (try list, get, or delete)", sub)
				},
			},
			{
				Name:  "search",
				Usage: "Search documents: search <collection> -q <query> --query-by <fields>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "q", Value: "*", Usage: "Search query string"},
					&cli.StringFlag{Name: "query-by", Value: "", Usage: "Comma-separated fields to query"},
					&cli.StringFlag{Name: "filter-by", Usage: "Filter expression (e.g., status:=published)"},
					&cli.StringFlag{Name: "sort-by", Usage: "Sorting expression (e.g., ratings_count:desc)"},
					&cli.IntFlag{Name: "per-page", Value: 10, Usage: "Number of results to return"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					colName := cmd.Args().First()
					if colName == "" {
						return fmt.Errorf("collection name is required. Usage: search <collection> -q <query>")
					}

					// Instantiate a dynamic DocumentsService scoped to map[string]any for unpredictable schemas
					docSvc := documents.NewDocumentsService[map[string]any](client, colName)
					params := schemas.SearchParams{
						Q:        cmd.String("q"),
						QueryBy:  cmd.String("query-by"),
						FilterBy: cmd.String("filter-by"),
						SortBy:   cmd.String("sort-by"),
						PerPage:  int(cmd.Int("per-page")),
					}

					res, err := docSvc.Search(ctx, params)
					if err != nil {
						return err
					}

					// Flatten hits for table rendering
					var docs []map[string]any
					for _, hit := range res.Hits {
						docs = append(docs, hit.Document)
					}
					printOutput(docs)
					return nil
				},
			},
			{
				Name:  "aliases",
				Usage: "Manage collection aliases: aliases [list|get <name>|delete <name>]",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sub := cmd.Args().First()
					if sub == "" || sub == "list" {
						res, err := client.Aliases.RetrieveAll(ctx)
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					if sub == "get" {
						res, err := client.Aliases.Retrieve(ctx, cmd.Args().Get(1))
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					if sub == "delete" {
						res, err := client.Aliases.Delete(ctx, cmd.Args().Get(1))
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					return fmt.Errorf("unknown subcommand: %s", sub)
				},
			},
			{
				Name:  "keys",
				Usage: "Manage API Keys: keys [list|get <id>|delete <id>]",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					sub := cmd.Args().First()
					if sub == "" || sub == "list" {
						res, err := client.Keys.RetrieveAll(ctx)
						if err != nil {
							return err
						}
						printOutput(res)
						return nil
					}
					return fmt.Errorf("unknown subcommand: %s", sub)
				},
			},
			{
				Name:  "health",
				Usage: "Check Typesense cluster health",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					res, err := client.Operations.Health(ctx)
					if err != nil {
						return err
					}
					printOutput(res)
					return nil
				},
			},
			{
				Name:  "metrics",
				Usage: "Check Typesense cluster metrics",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					res, err := client.Operations.Metrics(ctx)
					if err != nil {
						return err
					}
					printOutput(res)
					return nil
				},
			},
		},
	}
}

func printOutput(data any) {
	if expandedMode {
		renderExpanded(data)
	} else {
		renderTable(data)
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "collections", Description: "List, view or delete collections"},
		{Text: "search", Description: "Search documents in a collection"},
		{Text: "aliases", Description: "Manage collection aliases"},
		{Text: "keys", Description: "Manage API keys"},
		{Text: "health", Description: "Cluster health check"},
		{Text: "metrics", Description: "Typesense cluster metrics"},
		{Text: "\\x", Description: "Toggle expanded vertical output"},
		{Text: "\\config", Description: "Interactively update and encrypt credentials"},
		{Text: "\\q", Description: "Exit aql"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func parseArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		c := input[i]
		switch c {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if inQuotes {
				current.WriteByte(c)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// validateNodes tests connectivity to the /health endpoint of each provided node.
func validateNodes(nodes []string, apiKey string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	for _, n := range nodes {
		url := strings.TrimRight(n, "/") + "/health"

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("invalid URL format for %s: %v", n, err)
		}

		req.Header.Set("X-TYPESENSE-API-KEY", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %v", n, err)
		}

		_ = resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("node %s responded with status %d", n, resp.StatusCode)
		}
	}

	return nil
}

func cleanExit() {
	fmt.Println("Goodbye.")

	// Windows does not use stty, so we only run this on Unix-based systems
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "-raw", "echo")
		cmd.Stdin = os.Stdin
		_ = cmd.Run()
	}

	os.Exit(0)
}
