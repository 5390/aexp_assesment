// Package cli provides the Cobra-based CLI for inventory-cli.
package cli

import (
	"aexp_assesment/domain"
	"aexp_assesment/store"
	"aexp_assesment/util"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "inventory-cli",
		Short: "A product inventory management system",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// IMPORTANT: allow tests to inject store
			if productStore != nil {
				return nil
			}

			if cfg := viper.GetString("config"); cfg != "" {
				viper.SetConfigFile(cfg)
				if err := viper.ReadInConfig(); err != nil {
					return err
				}
			}

			lvlStr := strings.ToLower(viper.GetString("log-level"))
			lvl := slog.LevelInfo
			switch lvlStr {
			case "debug":
				lvl = slog.LevelDebug
			case "warn", "warning":
				lvl = slog.LevelWarn
			case "error":
				lvl = slog.LevelError
			}
			slog.SetDefault(slog.New(
				slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}),
			))

			var err error
			productStore, err = store.NewStore(
				viper.GetString("store"),
				viper.GetString("store-file"),
			)
			return err
		},
	}

	productStore domain.ProductStore
)

func init() {
	// shell
	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Interactive shell mode",
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
				rootCmd.SetArgs(strings.Fields(line))
				if err := rootCmd.Execute(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				rootCmd.SetArgs(nil)
			}
		},
	}
	rootCmd.AddCommand(shellCmd)

	rootCmd.PersistentFlags().String("store", "memory", "store backend: memory|file")
	rootCmd.PersistentFlags().String("store-file", "data/products.json", "file store path")
	rootCmd.PersistentFlags().String("config", "", "config file")
	rootCmd.PersistentFlags().String("log-level", "info", "log level")

	viper.BindPFlag("store", rootCmd.PersistentFlags().Lookup("store"))
	viper.BindPFlag("store-file", rootCmd.PersistentFlags().Lookup("store-file"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetEnvPrefix("INVENTORY")
	viper.AutomaticEnv()

	// create
	var name, category string
	var price float64
	var quantity int
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a product",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("name required")
			}
			id := util.GenerateUUID()
			p := domain.Product{ID: id, Name: name, Price: price, Quantity: quantity, Category: category}
			start := time.Now()
			if err := productStore.Create(context.Background(), p); err != nil {
				slog.Error("create failed", "product_id", id, "error", err)
				return err
			}
			slog.Info("product created", "product_id", id, "duration_ms", time.Since(start).Milliseconds())
			b, _ := json.MarshalIndent(p, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "name")
	createCmd.Flags().Float64Var(&price, "price", 0, "price")
	createCmd.Flags().IntVar(&quantity, "quantity", 0, "quantity")
	createCmd.Flags().StringVar(&category, "category", "", "category")
	rootCmd.AddCommand(createCmd)

	// get
	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get product by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := productStore.Get(context.Background(), args[0])
			if err != nil {
				if domain.IsProductNotFoundError(err) {
					fmt.Fprintln(os.Stderr, err)
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

	// update
	var uName, uCategory string
	var uPrice float64
	var uQuantity int
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

			if err := domain.ValidateProduct(p); err != nil {
				return err
			}

			start := time.Now()
			if err := productStore.Update(context.Background(), id, p); err != nil {
				slog.Error("update failed", "product_id", id, "error", err)
				return err
			}

			slog.Info(
				"product updated",
				"product_id", id,
				"duration_ms", time.Since(start).Milliseconds(),
			)

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

	// list
	var lCategory, lSort, lOrder, lOutput string
	var lMin, lMax float64
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
			out, err := productStore.List(context.Background(), domain.ListFilter{
				Category: lCategory,
				MinPrice: minPtr,
				MaxPrice: maxPtr,
				SortBy:   lSort,
				Order:    lOrder,
			})
			if err != nil {
				return err
			}
			if lOutput == "json" {
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			for _, p := range out {
				fmt.Printf("%s | %s | %.2f | %d | %s\n",
					p.ID, p.Name, p.Price, p.Quantity, p.Category)
			}
			return nil
		},
	}
	listCmd.Flags().StringVar(&lCategory, "category", "", "category")
	listCmd.Flags().Float64Var(&lMin, "min-price", 0, "min price")
	listCmd.Flags().Float64Var(&lMax, "max-price", 0, "max price")
	listCmd.Flags().StringVar(&lSort, "sort-by", "", "sort field")
	listCmd.Flags().StringVar(&lOrder, "order", "asc", "sort order")
	listCmd.Flags().StringVar(&lOutput, "output", "", "output format")
	rootCmd.AddCommand(listCmd)

	// delete
	var force bool
	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Delete %s? (y/N): ", args[0])
				var resp string
				if _, err := fmt.Scanln(&resp); err != nil || (resp != "y" && resp != "Y") {
					fmt.Println("aborted")
					return nil
				}
			}
			if err := productStore.Delete(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Println("deleted")
			return nil
		},
	}
	deleteCmd.Flags().BoolVar(&force, "force", false, "skip confirmation")
	rootCmd.AddCommand(deleteCmd)

	// import (FIXED: supports NDJSON)
	var importFile string
	importCmd := &cobra.Command{
		Use:   "import --file <file>",
		Short: "Import products from JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			if importFile == "" {
				return errors.New("--file required")
			}

			b, err := os.ReadFile(importFile)
			if err != nil {
				return err
			}

			btrim := bytes.TrimSpace(b)
			if len(btrim) == 0 {
				return errors.New("empty file")
			}

			var products []domain.Product

			// JSON array
			if btrim[0] == '[' {
				if err := json.Unmarshal(btrim, &products); err != nil {
					return err
				}
			} else {
				// NDJSON or single JSON object
				scanner := bufio.NewScanner(bytes.NewReader(btrim))
				for scanner.Scan() {
					line := bytes.TrimSpace(scanner.Bytes())
					if len(line) == 0 {
						continue
					}
					var p domain.Product
					if err := json.Unmarshal(line, &p); err != nil {
						return err
					}
					products = append(products, p)
				}
				if err := scanner.Err(); err != nil {
					return err
				}
			}

			return productStore.BulkImport(context.Background(), products)
		},
	}
	importCmd.Flags().StringVar(&importFile, "file", "", "input file")
	rootCmd.AddCommand(importCmd)

	// export
	var exportFile, exportCategory string
	exportCmd := &cobra.Command{
		Use:   "export --file <file>",
		Short: "Export products to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			if exportFile == "" {
				return errors.New("--file required")
			}
			out, err := productStore.List(context.Background(), domain.ListFilter{
				Category: exportCategory,
			})
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return os.WriteFile(exportFile, b, 0o644)
		},
	}
	exportCmd.Flags().StringVar(&exportFile, "file", "", "output file")
	exportCmd.Flags().StringVar(&exportCategory, "category", "", "category")
	rootCmd.AddCommand(exportCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
