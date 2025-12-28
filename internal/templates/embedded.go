package templates

import "embed"

//go:embed auth.yaml ecommerce.yaml content.yaml graphql-api.yaml mobile-api.yaml gateway.yaml node/**/*.tmpl infra/**/*.tmpl go/**/*.tmpl python/**/*.tmpl
var TemplateFS embed.FS
