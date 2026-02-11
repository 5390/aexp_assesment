// Package cli defines the command-line interface for the inventory-cli application.
//
// This package builds a Cobra-based CLI providing commands to manage products
// (create/get/list/update/delete/import/export) and an interactive `shell`
// mode. The commands operate on a `domain.ProductStore` which can be an in-memory
// or file-backed implementation.
package cli

import (
	"aexp_assesment/domain"
	"aexp_assesment/store"
	"aexp_assesment/util"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"bufio"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// rootCmd is the top-level Cobra command for the CLI. Persistent flags
	// (like storage backend selection) are defined here and a
	// PersistentPreRunE hook initializes the chosen `domain.ProductStore` once.
	rootCmd = &cobra.Command{
		Use:   "inventory-cli",
		Short: "A product inventory management system",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// initialize the store only once (avoid recreating on every command)
			if productStore != nil {
				return nil
			}
			kind, _ := cmd.Flags().GetString("store")
			path, _ := cmd.Flags().GetString("store-file")
			var err error
			productStore, err = store.NewStore(kind, path)
			return err
		},
	}
	// productStore is the currently-initialized domain.ProductStore instance used by
	// commands. It is configured by persistent flags and created once by
	// `PersistentPreRunE` above.
	productStore domain.ProductStore
)

// init registers all Cobra subcommands and their flags. Each command's
// behavior is implemented inline using `RunE` handlers that call into the
// `domain.ProductStore` interface.
func init() {
	// shell (interactive)
	// shellCmd starts an interactive REPL where users can enter commands
	// repeatedly without restarting the binary. It uses simple whitespace
	// splitting for arguments (does not handle quoted strings).
	//
	// Example:
	//   inventory> create --name "Laptop" --price 999.99
	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Interactive shell mode (type 'exit' or 'quit' to leave)",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("inventory> ")
				line, err := r.ReadString('\n')
				if err != nil {
					return nil
				}
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if line == "exit" || line == "quit" {
					return nil
				}
				// naive split (doesn't handle quotes)
				parts := strings.Fields(line)
				// set args and execute
				rootCmd.SetArgs(parts)
				if err := rootCmd.Execute(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				// clear args for next iteration
				rootCmd.SetArgs(nil)
			}
		},
	}
	rootCmd.AddCommand(shellCmd)

	rootCmd.PersistentFlags().String("store", "memory", "store backend: memory|file")
	rootCmd.PersistentFlags().String("store-file", "data/products.json", "file path for file store")

	// create
	// createCmd creates a new Product with a generated UUID and validates
	// the required fields before inserting into the configured store.
	var name string
	var price float64
	var quantity int
	var category string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new product",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("name required")
			}
			id := util.GenerateUUID()
			p := domain.Product{ID: id, Name: name, Price: price, Quantity: quantity, Category: category}
			if err := productStore.Create(context.Background(), p); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(p, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "product name")
	createCmd.Flags().Float64Var(&price, "price", 0, "product price")
	createCmd.Flags().IntVar(&quantity, "quantity", 0, "product quantity")
	createCmd.Flags().StringVar(&category, "category", "", "product category")
	rootCmd.AddCommand(createCmd)

	// get
	// getCmd retrieves a product by id and prints it as JSON. If the
	// product is not found, it prints a friendly message to stderr.
	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get product by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			p, err := productStore.Get(context.Background(), id)
			if err != nil {
				if domain.IsProductNotFoundError(err) {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					return nil
				}
				return err
			}
			b, _ := json.MarshalIndent(p, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	rootCmd.AddCommand(getCmd)

	// list
	// listCmd lists products with optional filtering, sorting and JSON
	// output support. Filters are applied via flags.
	var lCategory string
	var lMin float64
	var lMax float64
	var lSort string
	var lOrder string
	var lOutput string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List products",
		RunE: func(cmd *cobra.Command, args []string) error {
			var minPtr, maxPtr *float64
			if cmd.Flags().Changed("min-price") {
				minPtr = &lMin
			}
			if cmd.Flags().Changed("max-price") {
				maxPtr = &lMax
			}
			out, err := productStore.List(context.Background(), domain.ListFilter{Category: lCategory, MinPrice: minPtr, MaxPrice: maxPtr, SortBy: lSort, Order: lOrder})
			if err != nil {
				return err
			}
			if lOutput == "json" {
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			for _, p := range out {
				fmt.Printf("%s | %s | %.2f | %d | %s\n", p.ID, p.Name, p.Price, p.Quantity, p.Category)
			}
			return nil
		},
	}
	listCmd.Flags().StringVar(&lCategory, "category", "", "filter by category")
	listCmd.Flags().Float64Var(&lMin, "min-price", 0, "min price")
	listCmd.Flags().Float64Var(&lMax, "max-price", 0, "max price")
	listCmd.Flags().StringVar(&lSort, "sort-by", "", "sort by: name|price|quantity")
	listCmd.Flags().StringVar(&lOrder, "order", "asc", "order: asc|desc")
	listCmd.Flags().StringVar(&lOutput, "output", "", "output format: json")
	rootCmd.AddCommand(listCmd)

	// update
	// updateCmd supports partial updates via flags. It loads the existing
	// product, applies changed flags, validates and writes the update.
	var uName string
	var uPrice float64
	var uQuantity int
	var uCategory string
	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			p, err := productStore.Get(context.Background(), id)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("name") {
				p.Name = uName
			}
			if cmd.Flags().Changed("price") {
				p.Price = uPrice
			}
			if cmd.Flags().Changed("quantity") {
				p.Quantity = uQuantity
			}
			if cmd.Flags().Changed("category") {
				p.Category = uCategory
			}
			if err := productStore.Update(context.Background(), id, p); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(p, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	updateCmd.Flags().StringVar(&uName, "name", "", "name")
	updateCmd.Flags().Float64Var(&uPrice, "price", 0, "price")
	updateCmd.Flags().IntVar(&uQuantity, "quantity", 0, "quantity")
	updateCmd.Flags().StringVar(&uCategory, "category", "", "category")
	rootCmd.AddCommand(updateCmd)

	// delete
	// deleteCmd removes a product by id. By default it prompts for
	// confirmation; `--force` skips the prompt.
	var force bool
	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !force {
				fmt.Printf("Delete %s? (y/N): ", id)
				var resp string
				_, err := fmt.Scanln(&resp)
				if err != nil || (resp != "y" && resp != "Y") {
					fmt.Println("aborted")
					return nil
				}
			}
			if err := productStore.Delete(context.Background(), id); err != nil {
				return err
			}
			fmt.Println("deleted")
			return nil
		},
	}
	deleteCmd.Flags().BoolVar(&force, "force", false, "force delete without confirmation")
	rootCmd.AddCommand(deleteCmd)

	// import
	// importCmd loads products from a JSON file and performs a bulk import.
	// Supported formats: JSON array, single JSON object, or newline-delimited
	// JSON (NDJSON). The command validates the file and delegates to
	// `domain.ProductStore.BulkImport` for concurrent processing.
	var importFile string
	importCmd := &cobra.Command{
		Use:   "import --file <file>",
		Short: "Import products from JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if importFile == "" {
				return errors.New("--file required")
			}
			b, err := os.ReadFile(importFile)
			if err != nil {
				return err
			}
			btrim := bytes.TrimLeftFunc(b, func(r rune) bool { return r == ' ' || r == '\n' || r == '\t' || r == '\r' })
			var products []domain.Product
			if len(btrim) == 0 {
				return errors.New("empty import file")
			}
			switch btrim[0] {
			case '[':
				if err := json.Unmarshal(b, &products); err != nil {
					return err
				}
			case '{':
				// could be single object or NDJSON; try single object first
				var p domain.Product
				if err := json.Unmarshal(b, &p); err == nil {
					products = append(products, p)
				} else {
					// try NDJSON: decode line by line
					lines := bytes.Split(b, []byte{'\n'})
					for _, ln := range lines {
						ln = bytes.TrimSpace(ln)
						if len(ln) == 0 {
							continue
						}
						var pi domain.Product
						if err := json.Unmarshal(ln, &pi); err != nil {
							return err
						}
						products = append(products, pi)
					}
				}
			default:
				return errors.New("unsupported JSON format for import")
			}

			if err := productStore.BulkImport(context.Background(), products); err != nil {
				return err
			}
			fmt.Printf("imported %d products\n", len(products))
			return nil
		},
	}
	importCmd.Flags().StringVar(&importFile, "file", "", "json file to import")
	rootCmd.AddCommand(importCmd)

	// export
	// exportCmd writes filtered products to a file as a JSON array.
	var exportFile string
	var exportCategory string
	exportCmd := &cobra.Command{
		Use:   "export --file <file>",
		Short: "Export products to JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if exportFile == "" {
				return errors.New("--file required")
			}
			var minPtr, maxPtr *float64
			out, err := productStore.List(context.Background(), domain.ListFilter{Category: exportCategory, MinPrice: minPtr, MaxPrice: maxPtr})
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			return os.WriteFile(exportFile, b, 0o644)
		},
	}
	exportCmd.Flags().StringVar(&exportFile, "file", "", "output file")
	exportCmd.Flags().StringVar(&exportCategory, "category", "", "optional category filter")
	rootCmd.AddCommand(exportCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
