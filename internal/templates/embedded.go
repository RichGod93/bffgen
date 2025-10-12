package templates

import "embed"

//go:embed auth.yaml ecommerce.yaml content.yaml node/**/*.tmpl infra/**/*.tmpl go/**/*.tmpl
var TemplateFS embed.FS
