package templates

import "embed"

//go:embed auth.yaml ecommerce.yaml content.yaml node/**/*.tmpl
var TemplateFS embed.FS

