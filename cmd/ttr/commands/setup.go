package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pinetoppeter/timetracker/internal/storage"
	"github.com/spf13/cobra"
)

func NewSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Initialize TimeTracker configuration and setup",
		Long:  "Initialize TimeTracker by creating configuration files.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🚀 Setting up TimeTracker...")

			// Create storage instance
			store, err := storage.NewStorage()
			if err != nil {
				fmt.Printf("❌ Error creating storage: %v\n", err)
				return
			}

			// Create config file if it doesn't exist
			configPath := filepath.Join(store.GetBaseDir(), "config.json")

			if _, err := os.Stat(configPath); err == nil {
				// File exists
				fmt.Println("✅ Configuration already exists")
			} else if os.IsNotExist(err) {
				// File doesn't exist - create it
				fmt.Println("📋 Creating configuration...")

				// Get system timezone
				loc := time.Now().Location().String()
				if loc == "" {
					loc = "UTC"
				}

				config := map[string]interface{}{
					"timezone": loc,
					// Removed rounding setting as requested
				}

				configJSON, err := json.MarshalIndent(config, "", "  ")
				if err != nil {
					fmt.Printf("⚠️  Warning: Could not create config: %v\n", err)
				} else {
					if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
						fmt.Printf("⚠️  Warning: Could not write config file: %v\n", err)
					} else {
						fmt.Printf("✅ Configuration created at: %s\n", configPath)
					}
				}
			} else {
				// Other error occurred
				fmt.Printf("⚠️  Warning: Could not check config file status: %v\n", err)
			}

			// Create export schema if it doesn't exist
			schemaPath := filepath.Join(store.GetBaseDir(), "export-schema.json")
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				fmt.Println("📋 Creating export schema...")
				
				// Create empty schema
				schema := storage.NewEmptyExportSchema()
				if err := store.SaveExportSchema(schema); err != nil {
					fmt.Printf("⚠️  Warning: Could not create export schema: %v\n", err)
				} else {
					fmt.Printf("✅ Export schema created at: %s\n", schemaPath)
				}
				
				// Create example schema with common columns
				examplePath := filepath.Join(store.GetBaseDir(), "export-schema-example.json")
				exampleSchema := storage.NewDefaultExportSchema()
				
				exampleContent, err := json.MarshalIndent(exampleSchema, "", "  ")
				if err != nil {
					fmt.Printf("⚠️  Warning: Could not create export schema example: %v\n", err)
				} else {
					if err := os.WriteFile(examplePath, exampleContent, 0644); err != nil {
						fmt.Printf("⚠️  Warning: Could not write export schema example: %v\n", err)
					} else {
						fmt.Printf("📋 Example export schema created at: %s\n", examplePath)
						fmt.Println("💡 You can copy this file to export-schema.json and customize it for your needs")
					}
				}
			} else {
				fmt.Println("✅ Export schema already exists")
			}

			fmt.Println("\n🎉 TimeTracker setup complete!")
			
			// Set up shell completion
			setupCompletion()
		},
	}

	return cmd
}

// setupCompletion handles shell completion setup for the setup command
func setupCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not determine home directory for completion setup: %v\n", err)
		return
	}

	completionPath := filepath.Join(homeDir, ".timetracker", "ttr_completion.bash")
	if _, err := os.Stat(completionPath); err == nil {
		fmt.Println("✅ Shell completion already set up")
		return
	}
	fmt.Println("DEBUG: Creating new completion script...")

	// Generate completion script using the custom compatible completion
	completionScript := `# ttr bash completion script - compatible with older bash versions
# This script provides basic completion without requiring bash-completion v2+
# UNIQUE MARKER: UPDATED COMPLETION SCRIPT FROM SETUP.GO

_ttr_complete() {
    local cur prev words cword
    
    # Use COMP_WORDS and COMP_CWORD which are more reliable
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Handle different completion contexts
    case "$prev" in
        ttr)
            # Complete top-level commands
            COMPREPLY=($(compgen -W "start pause resume end info export meta name switch setup completion help" -- "$cur"))
            return 0
            ;;
        start|name|switch|meta)
            # For start/name/switch/meta commands, complete record names from records directory
            # Try to get data folder from config, fall back to default locations
            local records_dir=""
            
            # Try to read data folder from config file
            local config_file="$HOME/.timetracker/config.json"
            if [ -f "$config_file" ]; then
                # Extract dataFolder from config JSON using grep/sed
                local data_folder=$(grep -o '"dataFolder"[[:space:]]*:[[:space:]]*"[^"]*"' "$config_file" 2>/dev/null | sed 's/.*"dataFolder"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
                if [ -n "$data_folder" ]; then
                    # Expand tilde if present
                    if [[ "$data_folder" == ~* ]]; then
                        data_folder="$HOME${data_folder:1}"
                    fi
                    records_dir="$data_folder/records"
                fi
            fi
            
            # Fall back to TTR_BASE_DIR if set (for testing)
            if [ -z "$records_dir" ] && [ -n "$TTR_BASE_DIR" ]; then
                records_dir="$TTR_BASE_DIR/records"
            fi
            
            # Final fallback to default location
            if [ -z "$records_dir" ]; then
                records_dir="$HOME/.timetracker/records"
            fi
            
            if [ -d "$records_dir" ]; then
                COMPREPLY=($(compgen -W "$(ls "$records_dir"/*.json 2>/dev/null | xargs -n1 basename -a | sed 's/\.json$//')" -- "$cur"))
            else
                # If no records directory, just show empty completion
                COMPREPLY=()
            fi
            return 0
            ;;
        export)
            # Complete export formats
            COMPREPLY=($(compgen -W "csv json" -- "$cur"))
            return 0
            ;;
        meta)
            # Complete meta subcommands
            COMPREPLY=($(compgen -W "add list" -- "$cur"))
            return 0
            ;;
        completion)
            # Complete completion types
            COMPREPLY=($(compgen -W "bash zsh fish powershell" -- "$cur"))
            return 0
            ;;
        *)
            # Default completion - show help or common options
            COMPREPLY=($(compgen -W "--help --version" -- "$cur"))
            return 0
            ;;
    esac
}

# Register the completion function
complete -o default -F _ttr_complete ttr
`

	// Write the completion script to file
	if err := os.WriteFile(completionPath, []byte(completionScript), 0644); err != nil {
		fmt.Printf("⚠️  Warning: Could not write completion script: %v\n", err)
		return
	}

	// Detect shell and get appropriate RC file
	shell := detectShell()
	shellRC := getShellRCFile(shell)

	// Add instructions to shell RC file
	if _, err := os.Stat(shellRC); err == nil {
		content, err := os.ReadFile(shellRC)
		if err == nil {
			completionLine := "source " + completionPath
			if !containsString(string(content), completionLine) {
				f, err := os.OpenFile(shellRC, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString("\n# TimeTracker shell completion\n")
					f.WriteString("source " + completionPath + "\n")
					f.Close()
					fmt.Println("✨ Shell completion has been set up!")
					fmt.Printf("💡 Please restart your %s shell or run: source %s\n", shell, shellRC)
				} else {
					fmt.Printf("⚠️  Warning: Could not modify %s: %v\n", shellRC, err)
				}
			} else {
				fmt.Println("✅ Shell completion already configured in RC file")
			}
		} else {
			fmt.Printf("⚠️  Warning: Could not read %s: %v\n", shellRC, err)
		}
	} else {
		fmt.Printf("⚠️  Warning: Could not find %s for completion setup\n", shellRC)
	}
}

// Helper functions for shell detection
func detectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	return "bash"
}

func getShellRCFile(shell string) string {
	homeDir, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc")
	case "bash":
		return filepath.Join(homeDir, ".bashrc")
	default:
		return filepath.Join(homeDir, ".bashrc")
	}
}

func containsString(content, substring string) bool {
	return strings.Contains(content, substring)
}
