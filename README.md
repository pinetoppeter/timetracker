# TimeTracker CLI 🕒

A comprehensive time tracking tool with shell autocompletion, metadata support, and CSV export.

## Features

- **Time Tracking**: Track time with sessions and records
- **Shell Autocompletion**: Bash and Zsh support for record names in `name`, `switch`, and `start` commands
- **Metadata Support**: Add custom metadata to records for better organization
- **CSV Export**: Export time data with customizable columns
- **Record Management**: Name, switch, and manage time records
- **JSON Schema**: Configurable export schema for custom columns

## Installation

```bash
# Clone the repository
git clone https://github.com/pinetoppeter/timetracker.git
cd timetracker

# Build the application
go build -o ttr ./cmd/ttr/main.go

# Install to your PATH (optional)
sudo mv ttr /usr/local/bin/
```

## First Time Setup

Run the setup command to initialize configuration:

```bash
ttr setup
```

This will:
- Create the configuration file (`~/.timetracker/config.json`) with timezone settings
- Create the export schema file (`~/.timetracker/export-schema.json`)
- Set up shell autocompletion automatically for Bash or Zsh

## Usage

### Basic Commands

```bash
# Start a new session
ttr start

# Name the current record
ttr name "Project Work"

# Switch to a new record
ttr switch "Meeting"

# End the current session
ttr end

# Show session information
ttr info

# Export time data to CSV
ttr export
```

### Advanced Features

#### Metadata Management

Add custom metadata to records for better organization and reporting:

```bash
# Add metadata to current record
ttr meta project "Client Website"
ttr meta client "Acme Corp"
ttr meta billable true
ttr meta priority 1

# Add metadata to specific record
ttr meta "record-name" project "Client Website"
ttr meta "record-name" client "Acme Corp"

# List current record metadata
ttr meta list

# List specific record metadata
ttr meta list "record-name"

# Metadata is stored in record JSON files
# You can also edit these files manually
```

#### Shell Completion

TimeTracker supports shell autocompletion for record names in the `name`, `switch`, `start`, and `meta` commands:

```bash
# Completion is set up automatically on first run!
ttr name <TAB>        # Shows available record names
ttr switch <TAB>      # Shows available record names
ttr start <TAB>       # Shows available record names
ttr meta <TAB>        # Shows available record names
```

**Automatic Setup:**
- On first run, TimeTracker automatically:
  - Detects your shell (Bash or Zsh)
  - Generates completion script in `~/.timetracker/ttr_completion.bash` (Bash) or `~/.timetracker/ttr_completion.zsh` (Zsh)
  - Adds sourcing command to your shell RC file (`~/.bashrc` or `~/.zshrc`)
  - Prints friendly setup instructions
- Just restart your shell or run `source ~/.bashrc` (or `source ~/.zshrc` for Zsh)

**Manual Setup (if needed):**
```bash
# For Bash
ttr completion bash > ~/.timetracker/ttr_completion.bash
echo "source ~/.timetracker/ttr_completion.bash" >> ~/.bashrc
source ~/.bashrc

# For Zsh
ttr completion zsh > ~/.timetracker/ttr_completion.zsh
echo "source ~/.timetracker/ttr_completion.zsh" >> ~/.zshrc
source ~/.zshrc
```

#### CSV Export

Export your time data with customizable columns:

```bash
# Export current month
ttr export

# Customize export columns by editing the schema
nano ~/.timetracker/export-schema.json

# Export includes:
# - Record names and durations
# - All metadata properties
# - Schema-defined columns with defaults
```

## Configuration

### Export Schema

The export schema defines which columns appear in CSV exports. By default, TimeTracker creates an empty schema that you can customize. An example file with common column definitions is also provided:

```bash
# Main export schema (empty by default):
# Location: ~/.timetracker/export-schema.json

# Example schema with common columns:
# Location: ~/.timetracker/export-schema-example.json

# To use the example schema, copy it to the main schema file:
cp ~/.timetracker/export-schema-example.json ~/.timetracker/export-schema.json

# Example schema content:
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "TimeTracker Export Configuration",
  "description": "Configuration for TimeTracker CSV export columns",
  "type": "object",
  "columns": [
    {
      "name": "project",
      "displayName": "Project",
      "description": "Project name or identifier",
      "default": "",
      "type": "string"
    },
    {
      "name": "client",
      "displayName": "Client",
      "description": "Client or customer name",
      "default": "Unassigned",
      "type": "string"
    },
    {
      "name": "billable",
      "displayName": "Billable",
      "description": "Whether the time is billable",
      "default": "false",
      "type": "boolean"
    }
  ],
  "schemaVersion": "1.0",
  "createdBy": "TimeTracker"
}
```

### Record Metadata

Records are stored as JSON files that you can manually edit:

```bash
# Location: YOUR_DATA_FOLDER/records/
# Format: {record-name}.json

# Example record file:
{
  "name": "project-work",
  "metadata": {
    "project": "Website Redesign",
    "client": "Acme Corp",
    "billable": true,
    "priority": 1
  }
}
```

**Note:** Replace `YOUR_DATA_FOLDER` with your actual data folder path (e.g., `~/.timetracker_data/`).

## File Structure

### Configuration Files (Always in ~/.timetracker/)

```
~/.timetracker/
├── config.json                    # Configuration with timezone settings
├── ttr_completion.bash            # Bash completion script (auto-generated)
└── ttr_completion.zsh             # Zsh completion script (auto-generated)
```

### Time Tracking Data (User-Selected Location)

**Important:** During setup, you choose where your time tracking data is stored. This is separate from the configuration directory.

```
YOUR_CUSTOM_DATA_FOLDER/ (e.g., ~/.timetracker_data/ or ~/Documents/TimeTracker/)
├── export-schema.json             # Export configuration (empty by default)
├── export-schema-example.json     # Example schema with common columns
├── records/                       # Individual record metadata files
│   ├── record1.json               # Record with optional metadata
│   └── record2.json               # Record with optional metadata
└── sessions/                      # Session data (organized by month)
    └── YYYY-MM/                    # Monthly folders
        ├── session-*.json         # Session metadata
        └── session-*.csv           # Session CSV export
```


## Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `ttr start` | Start a new time tracking session |
| `ttr end` | End the current session |
| `ttr pause` | End the current record (modified from original pause) |
| `ttr resume` | Resume the last named record |
| `ttr switch [name]` | Switch to a new record (with optional name) |
| `ttr name [name]` | Name the current record |
| `ttr info` | Show current session information |
| `ttr export` | Export time data to CSV |
| `ttr setup` | Initialize configuration files |

### Metadata Commands

| Command | Description |
|---------|-------------|
| `ttr meta key value` | Add metadata to current record |
| `ttr meta list` | List current record metadata |

### Completion Commands

Generate shell completion scripts manually if needed:

| Command | Description |
|---------|-------------|
| `ttr completion bash` | Generate Bash completion script |
| `ttr completion zsh` | Generate Zsh completion script |
| `ttr completion fish` | Generate Fish completion script |
| `ttr completion powershell` | Generate PowerShell completion script |

## Examples

### Basic Workflow

```bash
# Start work
ttr start

# Name your first task
ttr name "Project Planning"

# Add metadata
ttr meta project "Website Redesign"
ttr meta client "Acme Corp"

# Switch to a meeting
ttr switch "Client Meeting"

# Add meeting metadata
ttr meta project "Website Redesign"
ttr meta client "Acme Corp"
ttr meta billable true

# Back to development
ttr switch "Development Work"

# End session when done
ttr end

# Export your time
ttr export
```

### Monthly Reporting

```bash
# Export January data
ttr export
# (Follow prompts for month/year)

# Customize columns by editing schema
nano ~/.timetracker/export-schema.json

# Re-export with new columns
ttr export
```

## Troubleshooting

### Shell Completion Not Working

1. **Restart your shell** or run:
   ```bash
   # For Bash
   source ~/.bashrc
   
   # For Zsh
   source ~/.zshrc
   ```

2. **Check completion script location**:
   ```bash
   cat ~/.timetracker/ttr_completion.bash
   ```

3. **Verify shell RC setup**:
   ```bash
   # For Bash
   grep "ttr_completion" ~/.bashrc
   
   # For Zsh
   grep "ttr_completion" ~/.zshrc
   ```

4. **Manual setup** (if automatic setup failed):
   ```bash
   # Generate completion script
   ttr completion bash > ~/.timetracker/ttr_completion.bash
   
   # For Bash
   echo "source ~/.timetracker/ttr_completion.bash" >> ~/.bashrc
   source ~/.bashrc
   
   # For Zsh
   echo "source ~/.timetracker/ttr_completion.zsh" >> ~/.zshrc
   source ~/.zshrc
   ```

## Development

### Building

```bash
go build -o ttr ./cmd/ttr/main.go
```

### Testing

TimeTracker includes comprehensive test suites for all functionality.

#### Running All Tests (Containerized - Recommended)

TimeTracker uses **containerized testing** to protect your local environment. Tests run in isolated Docker containers and never touch your user files.

```bash
# Run all tests safely in a container
./run_tests.sh
```

**What happens:**
- Docker container is built with all dependencies
- Tests run in complete isolation
- Temporary directories are used for all test data
- Container is automatically cleaned up after tests
- Your local files remain untouched

**Requirements:**
- Docker must be installed and running
- No other dependencies needed on your system

#### Running Specific Tests

```bash
# Run a specific shell test
./run_tests.sh test_metadata_persistence

```

#### Test Coverage

**Go Unit Tests (3 packages):**
- `internal/config` - Configuration management
- `internal/session` - Session structures and state
- `internal/storage` - Record storage and metadata

**Shell Integration Tests (6 focused tests):**
- `test_autocomplete.sh` - Autocomplete functionality
- `test_metadata_persistence.sh` - Metadata across sessions
- `test_meta_export_integration.sh` - Metadata in exports
- `test_meta_name_completion_parity.sh` - Meta/name completion
- `test_meta_specific_record.sh` - Specific record metadata
- `test_end.sh` - Session ending

⚠️ **Important:** All testing must be done using the containerized approach (`./run_tests.sh`). Local testing is not supported and may modify your user files.

### Running the Application

```bash
# Run the application
ttr start
```

### Architecture

```
cmd/ttr/          # CLI commands
internal/        # Core functionality
  ├── config/     # Configuration management
  ├── record/     # Record management
  ├── session/    # Session management
  ├── storage/    # File storage
  └── setup/      # Setup and initialization
pkg/             # (Future) Libraries
```

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Support

For questions or issues, please open a GitHub issue.

---

**TimeTracker** - Simple, powerful time tracking for developers and teams. 🚀
