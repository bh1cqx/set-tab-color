# set-tab-color

A command-line tool to set iTerm2 tab, foreground, and background colors with support for profiles and multiple color formats.

## Features

- Set tab, foreground, and background colors individually or together
- Support for hex colors (#f80, #ff8800), CSS color names (red, blue, etc.), and "default"
- Profile-based configuration using TOML files
- Flexible configuration file location

## Requirements

- iTerm2 with the `it2setcolor` utility installed
  - `it2setcolor` is part of iTerm2's shell integration utilities
  - Must be located at `~/.iterm2/it2setcolor`
  - Install iTerm2's shell integration: iTerm2 → Install Shell Integration
- Go 1.16+ for building from source

## Installation

```bash
go build -o set-tab-color
```

## Usage

### Basic Usage

```bash
# Set tab color only
./set-tab-color -tab red

# Set foreground and background colors
./set-tab-color -fg white -bg black

# Set all three colors
./set-tab-color -tab #ff8800 -fg lightblue -bg darkgray

# Mix hex colors and CSS names
./set-tab-color -tab blue -fg #ffffff
```

### Profile Usage

```bash
# Use a predefined profile
./set-tab-color -profile development
```

## Configuration

### Configuration File Location

The configuration file is located at:
- `~/.config/set-tab-color.toml` (default)
- Or the path specified by the `SET_TAB_COLOR_CONFIG` environment variable

### Profile Format

Profiles are defined in TOML format. Each profile can specify any combination of tab, foreground, and background colors.

#### Example Configuration File

```toml
[profiles.development]
tab = "blue"
fg = "white"
bg = "black"

[profiles.production]
tab = "red"
fg = "yellow"

[profiles.staging]
tab = "#ff8800"
fg = "lightblue"
bg = "darkgray"

[profiles.minimal]
tab = "green"

[profiles.reset]
tab = "default"
fg = "default"
bg = "default"
```

#### Profile Properties

- `tab`: Tab color (optional)
- `fg`: Foreground/text color (optional)
- `bg`: Background color (optional)

Any combination of these properties can be specified in a profile. Unspecified colors will remain unchanged.

### Supported Color Formats

1. **Hex Colors**
   - Short form: `#f80` (expands to `#ff8800`)
   - Long form: `#ff8800`
   - Without hash: `ff8800`

2. **CSS Color Names**
   - Standard names: `red`, `blue`, `green`, `white`, `black`
   - Extended names: `lightblue`, `darkgray`, `orange`, etc.

3. **Special Values**
   - `default`: Restore default color

## Examples

### Command Line Examples

```bash
# Development environment - blue tab, white text on black background
./set-tab-color -profile development

# Production warning - red tab with yellow text
./set-tab-color -profile production

# Quick color change
./set-tab-color -tab green -fg white

# Reset to defaults
./set-tab-color -profile reset
```

### Configuration Examples

```toml
# Themed profiles for different environments
[profiles.dark-theme]
tab = "purple"
fg = "lightgray"
bg = "black"

[profiles.light-theme]
tab = "blue"
fg = "black"
bg = "white"

# Status-based profiles
[profiles.error]
tab = "red"
fg = "white"

[profiles.success]
tab = "green"
fg = "white"

[profiles.warning]
tab = "orange"
fg = "black"
```

## Error Handling

The tool will return appropriate error messages for:
- Invalid color formats
- Missing profiles
- Missing `it2setcolor` binary
- Configuration file syntax errors
- Mixing profile and individual color flags

## Environment Variables

- `SET_TAB_COLOR_CONFIG`: Override the default configuration file location
- `HOME`: Used to locate the default config directory and `it2setcolor` binary
