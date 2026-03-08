package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pinetoppeter/timetracker/cmd/ttr/commands"
	"github.com/pinetoppeter/timetracker/internal/setup"
	"github.com/spf13/cobra"
)

func detectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		if strings.Contains(shell, "zsh") {
			return "zsh"
		} else if strings.Contains(shell, "bash") {
			return "bash"
		}
	}

	// Check if we're running in bash
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}

	// Check if we're running in zsh
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}

	return "bash" // Default to bash
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

func setupCompletionIfNeeded() {
	setupCompletionWithFeedback(false)
}

func setupCompletionWithFeedback(showFeedback bool) bool {
	// Check if completion is already set up
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if showFeedback {
			fmt.Println("⚠️  Warning: Could not determine home directory for completion setup")
		}
		return false
	}

	// Detect shell to determine which completion script to use
	shell := detectShell()
	completionPath := ""
	completionScript := ""
	
	switch shell {
	case "zsh":
		completionPath = filepath.Join(homeDir, ".timetracker", "ttr_completion.zsh")
		// Generate zsh completion script
		completionScript = generateZshCompletionScriptContent()
	case "bash", "":
		completionPath = filepath.Join(homeDir, ".timetracker", "ttr_completion.bash")
		// Generate bash completion script
		completionScript = generateBashCompletionScriptContent()
	default:
		// For other shells, don't set up completion
		return false
	}

	// Create .timetracker directory if it doesn't exist
	timetrackerDir := filepath.Join(homeDir, ".timetracker")
	if err := os.MkdirAll(timetrackerDir, 0755); err != nil {
		if showFeedback {
			fmt.Println("⚠️  Warning: Could not create .timetracker directory for completion")
		}
		return false
	}

	// Create completion file
	file, err := os.Create(completionPath)
	if err != nil {
		if showFeedback {
			fmt.Println("⚠️  Warning: Could not create completion script file")
		}
		return false
	}
	defer file.Close()
	
	file.WriteString(completionScript)

	// Get appropriate RC file
	shellRC := getShellRCFile(shell)

	// Add instructions to shell RC file
	completionWasSetup := false
	if _, err := os.Stat(shellRC); err == nil {
		// Check if completion is already in the RC file
		content, err := os.ReadFile(shellRC)
		if err == nil {
			completionLine := "source " + completionPath
			if !containsString(string(content), completionLine) {
				f, err := os.OpenFile(shellRC, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString("\n# TimeTracker shell completion\n")
					f.WriteString("source " + completionPath + "\n")
					f.Close()
					completionWasSetup = true
				} else if showFeedback {
					fmt.Printf("⚠️  Warning: Could not modify %s: %v\n", shellRC, err)
				}
			}
		} else if showFeedback {
			fmt.Printf("⚠️  Warning: Could not read %s: %v\n", shellRC, err)
		}
	} else if showFeedback {
		fmt.Printf("⚠️  Warning: Could not find %s for completion setup\n", shellRC)
	}

	// Print user-friendly message if completion was just set up
	if completionWasSetup && showFeedback {
		fmt.Println("✨ Shell completion has been set up!")
		fmt.Printf("💡 Please restart your %s shell or run: source %s\n", shell, shellRC)
	}

	return completionWasSetup
}

func generateBashCompletionScriptContent() string {
	return `# ttr bash completion script - compatible with older bash versions
# This script provides basic completion without requiring bash-completion v2+
# UNIQUE MARKER: UPDATED COMPLETION SCRIPT FROM MAIN.GO

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
}

func generateZshCompletionScriptContent() string {
	completionScript := `#compdef ttr
# ttr zsh completion script
# This script provides completion for zsh
# UNIQUE MARKER: UPDATED COMPLETION SCRIPT FROM MAIN.GO

# This completion script requires zsh and the completion system to be loaded
# If completion doesn't work, ensure you have these lines in your .zshrc:
#   autoload -U +X compinit && compinit
#   autoload -U compdef

_ttr_complete() {
    local -a commands
    commands=(
        'start'
        'pause'
        'resume'
        'end'
        'info'
        'export'
        'meta'
        'name'
        'switch'
        'setup'
        'completion'
        'help'
    )
    
    local curcontext="${curcontext}" state line
    typeset -A opt_args
    
    _arguments \
        '1: :->command' \
        '*:: :->args'
    
    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case $line[1] in
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
                        local -a records
                        records=(${records_dir}/*.json(N:t:r))
                        _describe 'record' records
                    fi
                    ;;
                export)
                    local -a formats
                    formats=(
                        'csv'
                        'json'
                    )
                    _describe 'format' formats
                    ;;
                completion)
                    local -a types
                    types=(
                        'bash'
                        'zsh'
                        'fish'
                        'powershell'
                    )
                    _describe 'type' types
                    ;;
                *)
                    _message 'unknown command'
                    ;;
            esac
            ;;
    esac
}

compdef _ttr_complete ttr
`

	return completionScript
}

func containsString(content, substring string) bool {
	return strings.Contains(content, substring)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ttr",
		Short: "TimeTracker - CLI tool for tracking work time",
		Long: `TimeTracker is a CLI tool that lets you track the time you work on different projects during the day.

Complete documentation is available at https://github.com/pinetoppeter/timetracker`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add commands
	nameCmd := commands.NewNameCmd()
	rootCmd.AddCommand(commands.NewStartCmd())
	rootCmd.AddCommand(commands.NewPauseCmd())
	rootCmd.AddCommand(commands.NewSwitchCmd())
	rootCmd.AddCommand(commands.NewResumeCmd())
	rootCmd.AddCommand(commands.NewInfoCmd())
	rootCmd.AddCommand(nameCmd)
	rootCmd.AddCommand(commands.NewEndCmd())
	rootCmd.AddCommand(commands.NewExportCmd())
	rootCmd.AddCommand(commands.NewMetaCmd())
	rootCmd.AddCommand(commands.NewSetupCmd())
	rootCmd.AddCommand(commands.NewCompletionCmd())

	// Enable shell completion for record names
	rootCmd.CompletionOptions.DisableDefaultCmd = false

	// Check if we're running the completion command - if so, skip setup
	isCompletionCmd := false
	if len(os.Args) > 1 && os.Args[1] == "completion" {
		isCompletionCmd = true
	}

	// Check and ensure application setup before running any commands (except completion)
	if !isCompletionCmd {
		if err := setup.EnsureSetup(); err != nil {
			fmt.Printf("Setup error: %v\n", err)
			os.Exit(1)
		}
		// Set up shell completion automatically after setup is complete
		setupCompletionWithFeedback(false)
	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
