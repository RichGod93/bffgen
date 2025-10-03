package templates

import "embed"

//go:embed auth.yaml ecommerce.yaml content.yaml
var TemplateFS embed.FS

