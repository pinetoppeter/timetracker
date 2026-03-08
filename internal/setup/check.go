package setup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pinetoppeter/timetracker/internal/config"
	"github.com/pinetoppeter/timetracker/internal/storage"
)

// EnsureSetup checks if the application is properly set up and prompts for setup if needed
func EnsureSetup() error {
	// Try to load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		// Config doesn't exist, run full setup
		return RunInteractiveSetup()
	}

	// Check if data folder is configured
	if cfg.DataFolder == "" {
		// Prompt user for data folder
		return PromptForDataFolder()
	}

	// Check if data folder exists
	if !fileExists(cfg.DataFolder) {
		return fmt.Errorf("configured data folder %s does not exist", cfg.DataFolder)
	}

	return nil
}

// RunInteractiveSetup runs the full interactive setup process
func RunInteractiveSetup() error {
	fmt.Println("🚀 Setting up TimeTracker...")

	// Get config path
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("could not get config path: %w", err)
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		// Config exists, just need to add data folder
		return PromptForDataFolder()
	}

	// Create default config with timezone
	configObj, err := config.CreateDefaultConfig(configPath)
	if err != nil {
		return fmt.Errorf("could not create default config: %w", err)
	}
	_ = configObj // Use the config object to avoid unused variable error

	fmt.Printf("✅ Configuration created at: %s\n", configPath)

	// Prompt for data folder first
	if err := PromptForDataFolder(); err != nil {
		return fmt.Errorf("could not configure data folder: %w", err)
	}

	// Now create storage instance with the configured data folder
	store, err := storage.NewStorage()
	if err != nil {
		return fmt.Errorf("could not create storage: %w", err)
	}

	// Create empty export schema in the data folder
	schemaPath := store.GetExportSchemaPath()
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Create empty schema
		schema := storage.NewEmptyExportSchema()
		if err := store.SaveExportSchema(schema); err != nil {
			fmt.Printf("⚠️  Warning: Could not create export schema: %v\n", err)
		} else {
			fmt.Printf("✅ Export schema created at: %s\n", schemaPath)
		}
		
		// Create example schema with common columns
		examplePath := store.GetExportSchemaExamplePath()
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
	}

	// Set up shell completion automatically
	if err := setupCompletion(); err != nil {
		fmt.Printf("⚠️  Warning: Could not set up shell completion: %v\n", err)
		fmt.Println("💡 You can manually set up completion by running:")
		fmt.Println("   ttr completion bash > ~/.timetracker/ttr_completion.bash")
		fmt.Println("   echo 'source ~/.timetracker/ttr_completion.bash' >> ~/.bashrc")
		fmt.Println("   source ~/.bashrc")
		fmt.Println("\n   For Zsh:")
		fmt.Println("   ttr completion zsh > ~/.timetracker/ttr_completion.zsh")
		fmt.Println("   echo 'source ~/.timetracker/ttr_completion.zsh' >> ~/.zshrc")
		fmt.Println("   source ~/.zshrc")
	} else {
		fmt.Println("🎉 TimeTracker setup complete!")
	}

	return nil
}

// PromptForDataFolder asks the user to specify a data folder location
func PromptForDataFolder() error {
	fmt.Println("📁 TimeTracker needs a folder to store your time tracking data")
	fmt.Println("   (records, sessions, and CSV exports)")
	fmt.Println()

	// Suggest default location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}
	
	defaultDataFolder := filepath.Join(homeDir, ".timetracker_data")
	
	// Show suggested common locations
	fmt.Println("💡 Common location suggestions:")
	suggestions := []string{
		defaultDataFolder,
		filepath.Join(homeDir, "Documents", "TimeTracker"),
		filepath.Join(homeDir, "timetracker_data"),
		filepath.Join(homeDir, ".local", "share", "timetracker"),
	}
	
	for i, suggestion := range suggestions {
		fmt.Printf("  %d. %s\n", i+1, suggestion)
	}
	
	fmt.Println()
	fmt.Printf("Suggested location: %s\n", defaultDataFolder)
	fmt.Println("Enter data folder path (or press Enter to use suggested location):")
	fmt.Println("  - Type a number (1-4) to choose from suggestions above")
	fmt.Println("  - Type a custom path (e.g., /custom/path or ~/relative/path)")
	fmt.Println("  - Press Enter to accept the suggested location")
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("could not read input: %w", err)
	}

	// Trim whitespace
	dataFolder := strings.TrimSpace(input)
	
	// Check if user entered a number to choose from suggestions
	if choice, err := strconv.Atoi(dataFolder); err == nil && choice >= 1 && choice <= len(suggestions) {
		dataFolder = suggestions[choice-1]
	} else if dataFolder == "" {
		// Use default if empty
		dataFolder = defaultDataFolder
	}

	// Expand tilde if present
	if strings.HasPrefix(dataFolder, "~") {
		dataFolder = filepath.Join(homeDir, dataFolder[1:])
	}

	// Ensure absolute path
	if !filepath.IsAbs(dataFolder) {
		dataFolder = filepath.Join(homeDir, dataFolder)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dataFolder, 0755); err != nil {
		return fmt.Errorf("could not create data folder: %w", err)
	}

	// Load config and update data folder
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("could not get config path: %w", err)
	}

	var cfg *config.Config
	if _, err := os.Stat(configPath); err == nil {
		// Config exists, load it
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("could not load config: %w", err)
		}
	} else {
		// Create new config with default timezone
		cfg = &config.Config{
			Timezone: "UTC", // Will be set properly when saved
		}
		// Set timezone properly
		loc := time.Now().Location().String()
		if loc != "" {
			cfg.Timezone = loc
		}
	}

	// Update data folder
	cfg.DataFolder = dataFolder

	// Save updated config
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}

	fmt.Printf("✅ Data folder configured: %s\n", dataFolder)
	fmt.Println("🎉 TimeTracker setup complete!")

	return nil
}

// fileExists checks if a file or directory exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// detectShell detects the user's current shell
func detectShell() string {
	// First, try to detect from the current process's environment
	// This is the most reliable method when running interactively
	
	// Check if we're running in zsh by checking for ZSH_NAME first (most reliable)
	if os.Getenv("ZSH_NAME") != "" {
		return "zsh"
	}

	// Check for zsh-specific variables
	if os.Getenv("ZSH_VERSION") != "" || os.Getenv("ZDOTDIR") != "" {
		return "zsh"
	}

	// Check BASH environment variable
	if os.Getenv("BASH") != "" {
		return "bash"
	}

	// Check for bash-specific variables
	if os.Getenv("BASH_VERSION") != "" || os.Getenv("BASHOPTS") != "" {
		return "bash"
	}

	// Check SHELL environment variable as fallback
	shell := os.Getenv("SHELL")
	if shell != "" {
		if strings.Contains(shell, "zsh") {
			return "zsh"
		} else if strings.Contains(shell, "bash") {
			return "bash"
		}
	}

	// Fallback: try to detect from parent process tree
	if runtime.GOOS != "windows" {
		// On Unix-like systems, walk up the process tree to find the shell
		currentPID := os.Getppid()
		for currentPID > 1 {
			procCmdline := fmt.Sprintf("/proc/%d/cmdline", currentPID)
			if content, err := os.ReadFile(procCmdline); err == nil {
				cmdline := strings.ReplaceAll(string(content), "\x00", " ")
				cmdline = strings.TrimSpace(cmdline)
				// Look for shell names in the command line
				if strings.HasSuffix(cmdline, "/zsh") || strings.Contains(cmdline, " zsh") {
					return "zsh"
				} else if strings.HasSuffix(cmdline, "/bash") || strings.Contains(cmdline, " bash") {
					return "bash"
				}
			}
			// Move up to parent process
			currentPID = getParentPID(currentPID)
			if currentPID <= 1 {
				break
			}
		}
	}

	// If we still can't detect, try checking the current shell by running a command
	if runtime.GOOS != "windows" {
		if shell := detectShellByCommand(); shell != "" {
			return shell
		}
	}

	// Default to bash if we can't detect
	return "bash"
}

// detectShellByCommand tries to detect the shell by running a command
func detectShellByCommand() string {
	// Try to detect shell by checking what shell is running us
	cmd := exec.Command("ps", "-p", strconv.Itoa(os.Getppid()), "-o", "comm=")
	output, err := cmd.Output()
	if err == nil {
		shell := strings.TrimSpace(string(output))
		if strings.Contains(shell, "zsh") {
			return "zsh"
		} else if strings.Contains(shell, "bash") {
			return "bash"
		}
	}
	return ""
}

// setupCompletion adds completion script sourcing to shell RC files
func setupCompletion() error {
	shell := detectShell()
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("could not get config directory: %w", err)
	}

	completionScriptPath := ""
	sourceCommand := ""

	// Generate appropriate completion script based on shell type
	switch shell {
	case "zsh":
		completionScriptPath = filepath.Join(configDir, "ttr_completion.zsh")
		if err := generateZshCompletionScript(completionScriptPath); err != nil {
			return fmt.Errorf("could not generate zsh completion script: %w", err)
		}
	case "bash", "":
		completionScriptPath = filepath.Join(configDir, "ttr_completion.bash")
		if err := generateBashCompletionScript(completionScriptPath); err != nil {
			return fmt.Errorf("could not generate bash completion script: %w", err)
		}
	default:
		// For other shells, don't generate completion
		fmt.Println("🐚 Shell completion not supported for:", shell)
		return nil
	}

	fmt.Printf("✅ Completion script generated at: %s\n", completionScriptPath)

	// Add sourcing command to appropriate RC file
	rcFile := ""
	sourceCommand = fmt.Sprintf("source %s\n", completionScriptPath)

	switch shell {
	case "zsh":
		rcFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
		fmt.Println("🐚 Detected Zsh - setting up completion")
	case "bash", "":
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
		fmt.Println("🐚 Detected Bash - setting up completion")
	}

	if rcFile == "" {
		return fmt.Errorf("could not determine RC file for shell: %s", shell)
	}

	// Check if completion is already sourced
	if fileContainsString(rcFile, sourceCommand) {
		fmt.Printf("✅ Completion already configured in %s\n", rcFile)
		return nil
	}

	// Create backup of existing RC file if it exists and has content
	if fileExists(rcFile) {
		content, err := os.ReadFile(rcFile)
		if err == nil && len(content) > 0 {
			backupFile := rcFile + ".backup-" + time.Now().Format("20060102-150405")
			if err := os.WriteFile(backupFile, content, 0644); err == nil {
				// Backup created successfully (we don't announce this to avoid clutter)
				// This is just a safety measure
			}
		}
	}

	// Append source command to RC file
	file, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", rcFile, err)
	}
	defer file.Close()

	// Check current file size before writing
	if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
		// File has existing content, make sure we're at the end
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			return fmt.Errorf("could not seek to end of %s: %w", rcFile, err)
		}
	}

	if _, err := file.WriteString(sourceCommand); err != nil {
		return fmt.Errorf("could not write to %s: %w", rcFile, err)
	}

	fmt.Printf("✅ Added completion sourcing to %s\n", rcFile)
	fmt.Println("💡 Restart your shell or run:")
	fmt.Printf("   source %s\n", rcFile)

	return nil
}

// fileContainsString checks if a file contains a specific string
func fileContainsString(filename, search string) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), search)
}

// getParentPID reads the parent PID from /proc/[pid]/stat
func getParentPID(pid int) int {
	if pid <= 1 {
		return 0
	}
	
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	content, err := os.ReadFile(statPath)
	if err != nil {
		return 0
	}
	
	// Parse the stat file - format is: pid (comm) state ppid ...
	// We need to extract the 4th field (ppid)
	fields := strings.Fields(string(content))
	if len(fields) >= 4 {
		if ppid, err := strconv.Atoi(fields[3]); err == nil {
			return ppid
		}
	}
	return 0
}

// generateBashCompletionScript generates the bash completion script
func generateBashCompletionScript(path string) error {
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

	return os.WriteFile(path, []byte(completionScript), 0644)
}

// generateZshCompletionScript generates a zsh completion script
func generateZshCompletionScript(path string) error {
	completionScript := `# ttr zsh completion script
# This script provides completion for zsh

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

	return os.WriteFile(path, []byte(completionScript), 0644)
}