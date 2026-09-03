# AMARA

A robust, memory-efficient Typesense client and interactive REPL CLI powered by Go generics.

## Table of Contents

* [About the Project](#-about-the-project)
  * [Features](#-features)
  * [Tech Stack](#-tech-stack)
* [Getting Started](#-getting-started)
  * [Prerequisites](#prerequisites)
  * [Installation](#installation)
* [Usage](#-usage)
  * [Using the SDK](#using-the-sdk)
  * [Using the CLI (aql)](#using-the-cli-aql)
* [Configuration](#-configuration)
* [Adding Features & Contributing](#-adding-features--contributing)

## About the Project

This project provides a comprehensive Go integration for Typesense. It consists of two primary components designed for institutional-grade deployments:

1. **The SDK:** A modular Go client that interfaces with the Typesense API. It handles cluster node routing and failover via an atomic, lock-free `NodeManager`, implements context-aware retry policies, and utilizes Go generics (`[T any]`) to provide type-safe document operations. It also supports memory-efficient bulk imports via `io.Reader` streaming.
2. **The CLI (`aql`):** An interactive terminal shell built using `urfave/cli/v3` and `go-prompt`. It supports executing raw search queries, collection management, real-time metrics, and horizontal/vertical `\x` formatting, emulating `psql` styling.

### Features

* **Type-Safe Document Operations:** Leverage Go generics (`[T any]`) for strict typing when interacting with documents, or map to `map[string]any` for schemaless data.
* **Resilient Node Management:** Automatic failover, context-aware retries, and lock-free atomic health tracking via the `NodeManager`.
* **Memory-Efficient Imports:** Stream large JSONL datasets directly to Typesense via the `ImportStream` method without buffering entirely into memory.
* **Interactive CLI (`aql`):** A powerful REPL with auto-completion, tabular output, and a vertical expanded mode (`\x`) for deeply nested JSON.
* **12-Factor App Ready:** Configuration securely managed via environment variables with a fallback to local credential files.

### Tech Stack

* **Language:** Go (v1.26.2)
* **CLI Framework:** [urfave/cli/v3](https://github.com/urfave/cli)
* **Interactive Prompt:** [c-bata/go-prompt](https://github.com/c-bata/go-prompt)
* **Table Formatting:** [jedib0t/go-pretty/v6](https://github.com/jedib0t/go-pretty)
* **Testing:** [testify](https://github.com/stretchr/testify)

## Getting Started

### Prerequisites

* Go 1.21 or higher (Project uses Go 1.26.2 per `go.mod`)
* A running Typesense server or cluster


### Installation

You can install the `aql` CLI using one of the following methods:

#### Option 1: Pre-built Binaries (Recommended)
Our automated CI/CD pipeline attaches pre-compiled, standalone binaries for Linux (Arch, RHEL, Ubuntu, etc.), macOS, and Windows across both AMD64 and ARM64 architectures to every release.

1. Navigate to the [Releases page](https://github.com/Kynetic-Engynes-Platforms/amara/releases) of this repository.
2. Download the latest version corresponding to your operating system and architecture (e.g., `aql-linux-amd64`).
3. Make the binary executable and move it to your system's PATH:
   
```bash
   chmod +x aql-linux-amd64
   sudo mv aql-linux-amd64 /usr/local/bin/aql
```


#### Option 2: Build from Source
If you prefer to compile the application yourself, ensure you have Go installed on your system.

```bash
# Clone the repository
git clone [https://github.com/Kynetic-Engynes-Platforms/amara.git](https://github.com/Kynetic-Engynes-Platforms/amara.git)
cd amara

# Build the CLI binary
go build -o aql ./pkg/cmd/ 

# Make it globally accessible (optional)
sudo mv aql /usr/local/bin/
```


## Usage

### Using the SDK

You can integrate the SDK into your own Go applications to interact with Typesense using strict schemas.

```go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/connection"
	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types"
	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types/schemas"
)

// 1. Define your strict schema
type Book struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Views int    `json:"views"`
}

func main() {
	// 2. Initialize the client
	cfg := types.Config{
		APIKey: "your-api-key",
		Nodes:  []string{"http://localhost:8108"},
	}
	
	client, err := connection.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Connect to a Collection strictly typed to your struct
	booksCollection := connection.NewCollection[Book](client, "books")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 4. Perform a search
	searchParams := schemas.SearchParams{
		Q:       "golang",
		QueryBy: "title",
		PerPage: 10,
	}

	result, err := booksCollection.Documents.Search(ctx, searchParams)
	if err != nil {
		log.Fatal(err)
	}

	for _, hit := range result.Hits {
		fmt.Printf("Found: %s (Views: %d)\n", hit.Document.Title, hit.Document.Views)
	}
	
	// 5. Memory-Efficient Batch Streaming
	// Stream large datasets directly via io.Reader without holding them in memory.
	file, _ := os.Open("large_dataset.jsonl")
	defer file.Close()
	
	_, err = booksCollection.Documents.ImportStream(ctx, file, "upsert")
}
```


### Using the CLI (aql)

Launch the binary to enter the interactive shell:

```bash
./aql
```


Available Commands:


- `collections list`: List all collections in the cluster.

- `collections get <name>`: Retrieve schema for a specific collection.

- `collections delete <name>`: Delete a collection.

- `search <collection> -q "<query>" --query-by <fields>`: Search a collection (e.g., search books -q "golang" --query-by "title" --per-page 5).

- `aliases [list|get <name>|delete <name>]`: Manage collection aliases.

- `keys [list]`: Manage API keys.

- `health`: Check the cluster's health status.

- `metrics`: Display system metrics (CPU, RAM, Disk).

- `\x`: Toggle expanded (vertical) display mode—ideal for deeply nested documents (similar to psql).

- `\q` or `exit`: Quit the shell.


## Configuration

The client and CLI securely resolve configurations in the following order (adhering to 12-factor principles):

1. Environment Variables (Primary):

        * AMARA_SECURITY_KEY : 32-Byte key used for encryption and decryption.

		* AMARA_TYPESENSE_API_KEY: Your Typesense server API key.

        * AMARA_TYPESENSE_NODES: A comma-separated list of your node URLs (e.g., http://node1:8108,http://node2:8108).

2. Fallback Credentials File (creds.json):
    
    If environment variables are absent, the CLI attempts to read from ~/.amara/creds.json.