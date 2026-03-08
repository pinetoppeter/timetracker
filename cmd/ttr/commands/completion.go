package commands

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

// generateCompatibleBashCompletion generates a bash completion script
// that works with older bash versions without requiring bash-completion v2+
func generateCompatibleBashCompletion(rootCmd *cobra.Command) {
	// Header with compatible functions
	completionScript := `# ttr bash completion script - compatible with older bash versions
# This script provides basic completion without requiring bash-completion v2+

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

	// Write the completion script to stdout
	io.WriteString(os.Stdout, completionScript)
}

func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "completion [bash|zsh|fish|powershell]",
		Short:  "Generate the autocompletion script for ttr",
		Long:   "Generate the autocompletion script for ttr for the specified shell.",
		Hidden: true, // Hide this command from help
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add shell-specific completion subcommands
	cmd.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate Bash completion script",
		Run: func(cmd *cobra.Command, args []string) {
			// Generate a more compatible bash completion script
			generateCompatibleBashCompletion(cmd.Root())
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate Zsh completion script",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Root().GenZshCompletion(os.Stdout)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "fish",
		Short: "Generate Fish completion script",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Root().GenFishCompletion(os.Stdout, true)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		},
	})

	return cmd
}