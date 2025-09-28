# Regeneration-Safe Scaffolding

bffgen implements a comprehensive regeneration-safe scaffolding system that ensures idempotent writes, preserves user modifications, and provides powerful diffing capabilities.

## Features

### 🔒 Code Fence Markers

Generated code is wrapped in special markers that allow safe regeneration:

```go
// bffgen:begin
func generatedFunction() {
    fmt.Println("This is generated code")
}
// bffgen:end
```

**Custom Markers:**

```go
// bffgen:begin:routes
func routeHandlers() {
    // Route-specific generated code
}
// bffgen:end:routes
```

### 🔄 Idempotent Writes

- **Safe Regeneration**: Running `bffgen generate` multiple times produces identical results
- **User Code Preservation**: Code outside markers is never modified
- **Conflict Detection**: Overlapping changes are detected and reported

### 🔍 Dry-Run and Check Modes

**Check Mode (`--check`):**

```bash
bffgen generate --check
```

Shows what would be changed without making any modifications.

**Dry-Run Mode (`--dry-run`):**

```bash
bffgen generate --dry-run
```

Shows what would be changed with detailed diff output.

**Verbose Output (`--verbose`):**

```bash
bffgen generate --verbose
```

Provides detailed information about the generation process.

### 🔀 Three-Way Merge

When regenerating code, bffgen performs intelligent 3-way merging:

1. **Base**: Original generated content
2. **Local**: User's modifications
3. **Remote**: New generated content

The merge strategy:

- Preserves user modifications outside generated sections
- Updates generated content within markers
- Reports conflicts for overlapping changes

## Usage Examples

### Basic Generation

```bash
# Generate code with markers
bffgen generate

# Check what would change
bffgen generate --check

# Dry run with verbose output
bffgen generate --dry-run --verbose
```

### Regeneration with User Modifications

**Initial Generation:**

```go
package main

import "fmt"

// bffgen:begin
func generatedFunction() {
    fmt.Println("Generated code")
}
// bffgen:end
```

**User Adds Custom Code:**

```go
package main

import "fmt"

// bffgen:begin
func generatedFunction() {
    fmt.Println("Generated code")
}
// bffgen:end

// User added this function manually
func userFunction() {
    fmt.Println("This is user code")
}
```

**Regeneration Preserves User Code:**

```go
package main

import "fmt"

// bffgen:begin
func updatedGeneratedFunction() {
    fmt.Println("Updated generated code")
}
// bffgen:end

// User added this function manually
func userFunction() {
    fmt.Println("This is user code")
}
```

### Conflict Resolution

When conflicts occur, bffgen provides detailed information:

```
Merge Conflicts (1):
==================================================
Line 10:
  Base:   func oldFunction() {}
  Local:  func userModifiedFunction() {}
  Remote: func newGeneratedFunction() {}
```

## Implementation Details

### Marker System

The marker system uses regex patterns to identify generated sections:

```go
type CodeMarker struct {
    Begin string // "// bffgen:begin"
    End   string // "// bffgen:end"
}
```

### Diff Engine

The diff engine provides:

- **Line-by-line comparison**
- **Change type detection** (added, removed, modified)
- **Summary statistics**
- **Formatted output**

### Generator API

```go
generator := scaffolding.NewGenerator()
generator.SetCheckMode(true)
generator.SetDryRun(false)
generator.SetVerbose(true)

err := generator.GenerateFile("main.go", content)
```

## Best Practices

### 1. Always Use Markers

Wrap all generated code in markers:

```go
// bffgen:begin
// Generated code here
// bffgen:end
```

### 2. Test with Check Mode

Before running generation, use check mode:

```bash
bffgen generate --check
```

### 3. Preserve User Code

Never modify code outside of markers:

```go
// ✅ Good: User code outside markers
func userFunction() {
    // User code
}

// bffgen:begin
func generatedFunction() {
    // Generated code
}
// bffgen:end
```

### 4. Handle Conflicts Gracefully

When conflicts occur:

1. Review the conflict details
2. Choose the appropriate resolution
3. Test the merged result

## Advanced Features

### Custom Markers

Use custom markers for different types of generated content:

```go
// bffgen:begin:routes
func routeHandlers() {
    // Route-specific code
}
// bffgen:end:routes

// bffgen:begin:middleware
func middlewareHandlers() {
    // Middleware-specific code
}
// bffgen:end:middleware
```

### Backup System

Enable automatic backups:

```go
generator.SetBackupDir("./backups")
```

### Validation

Validate marker structure:

```go
err := generator.ValidateFile("main.go")
if err != nil {
    // Handle validation error
}
```

## Troubleshooting

### Common Issues

**1. Missing End Marker**

```
Error: unclosed section starting at line 5
```

**Solution**: Ensure all begin markers have corresponding end markers.

**2. Nested Markers**

```
Error: nested begin marker found at line 10
```

**Solution**: Don't nest markers within other markers.

**3. Merge Conflicts**

```
Merge Conflicts (1): Line 10
```

**Solution**: Review conflicts and choose appropriate resolution.

### Debug Mode

Enable verbose output for debugging:

```bash
bffgen generate --verbose
```

## Integration with bffgen

The regeneration-safe scaffolding is integrated into the `bffgen generate` command:

```bash
# Standard generation
bffgen generate

# Check mode
bffgen generate --check

# Dry run
bffgen generate --dry-run

# Verbose output
bffgen generate --verbose
```

This ensures that all generated code is safe to regenerate and preserves user modifications.
