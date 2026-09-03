package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/sdk"
	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli/v3"
)

var (
	client       *sdk.Client
	expandedMode bool // Tracks if the user has enabled '\x' expanded vertical display
)

func main() {
	// 1. CLI owns the credentials loading, keeping the SDK environment-agnostic.
	cfg, err := loadCLICredentials()
	if err != nil {
		fmt.Printf("Notice: Could not load creds.json, falling back to defaults (%v)\n", err)
		cfg = sdk.Config{
			Nodes:  []string{"http://localhost:8108"},
			APIKey: "xyz", // Default for local testing
		}
	}

	// 2. Initialize the pure SDK client
	c, err := sdk.NewClient(cfg)
	if err != nil {
		fmt.Printf("Fatal: Client initialization failed: %v\n", err)
		os.Exit(1)
	}
	client = c

	// 3. Start the interactive REPL
	fmt.Println("Typesense Shell (tsql) v1.0.0")
	fmt.Println("Type 'help' for available commands, '\\x' for expanded display, '\\q' or 'exit' to quit.")

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("tsql=> "),
		prompt.OptionTitle("tsql"),
	)
	p.Run()
}

// loadCLICredentials securely loads Typesense configuration exclusively for the terminal session.
func loadCLICredentials() (sdk.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return sdk.Config{}, err
	}

	path := filepath.Join(home, ".typesense-go", "creds.json")
	file, err := os.Open(path)
	if err != nil {
		return sdk.Config{}, err
	}
	defer file.Close()

	var cfg sdk.Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return sdk.Config{}, fmt.Errorf("invalid JSON in creds.json: %w", err)
	}

	return cfg, nil
}

// executor routes raw string input from the REPL prompt into the urfave CLI parser.
func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	// Intercept REPL-specific meta-commands (like psql)
	switch in {
	case "exit", "quit", "\\q":
		fmt.Println("Goodbye.")
		os.Exit(0)
	case "\\x":
		expandedMode = !expandedMode
		if expandedMode {
			fmt.Println("Expanded display is on.")
		} else {
			fmt.Println("Expanded display is off.")
		}
		return
	}

	args := append([]string{"tsql"}, parseArgs(in)...)
	cmd := buildCLI()

	if err := cmd.Run(context.Background(), args); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// buildCLI defines the robust command surface area for the terminal application.
func buildCLI() *cli.Command {
	return &cli.Command{
		Name:  "tsql",
		Usage: "Typesense production interactive CLI",
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
					docSvc := sdk.NewDocumentsService[map[string]any](client, colName)
					params := sdk.SearchParams{
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

// printOutput delegates formatting to printer.go depending on the current expandedMode state.
// (Note: Requires renderTable and renderVertical to be defined in printer.go)
func printOutput(data any) {
	if expandedMode {
		renderExpanded(data) // Defined in printer.go
	} else {
		renderTable(data) // Defined in printer.go
	}
}

// completer provides auto-complete suggestions for the go-prompt REPL.
func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "collections", Description: "List, view or delete collections"},
		{Text: "search", Description: "Search documents in a collection"},
		{Text: "aliases", Description: "Manage collection aliases"},
		{Text: "keys", Description: "Manage API keys"},
		{Text: "health", Description: "Cluster health check"},
		{Text: "metrics", Description: "Typesense cluster metrics"},
		{Text: "\\x", Description: "Toggle expanded vertical output"},
		{Text: "\\q", Description: "Exit tsql"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// parseArgs gracefully handles strings containing quoted phrases (e.g., search col -q "San Francisco").
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
