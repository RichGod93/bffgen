# Template Files - Ignore Syntax Errors

**Note:** Files in this directory contain Go template variables and will be invalid JavaScript/Go until processed by bffgen.

## Template Variable Syntax

- `{{PROJECT_NAME}}` - Project name
- `{{PORT}}` - Server port
- `{{ range .BackendServices }}...{{ end }}` - Loop over backend services
- `{{ .Name | ToPascalCase }}` - Transform functions

These variables are replaced during project generation. Ignore syntax errors from linters/IDEs.

## For Developers

When editing template files:
1. Ignore JavaScript/TypeScript/Go linter errors
2. Template variables use `{{VAR}}` syntax
3. Files are valid after bffgen processes them
4. Test by running `bffgen init` with the template
