# Template System Design

## Overview

The bffgen template system provides a flexible, extensible mechanism for creating project scaffolds. It supports three template sources and arbitrary variable substitution.

## Template Format Specification

### Manifest File (template.yaml)

Every template must include a `template.yaml` manifest in its root directory:

```yaml
name: my-template           # Required: Template identifier
version: 1.0.0              # Required: Semantic version
description: "..."          # Required: Human-readable description
author: username            # Required: Author/organization
category: web               # Required: web, mobile, microservices, fullstack
language: nodejs-express    # Required: Target runtime

features:                   # Optional: List of features
  - JWT Authentication
  - Rate Limiting
  - GraphQL Support

variables:                  # Optional: Template variables
  - name: PROJECT_NAME
    description: "Name of the project"
    default: my-project
    required: false
  
  - name: DATABASE_URL
    description: "Database connection string"
    required: true

files:                      # Optional: File patterns to include
  - src/**
  - config/**

post_install:               # Optional: Post-installation commands
  - npm install
  - echo "Setup complete!"
```

### Directory Structure

```
my-template/
├── template.yaml          # Manifest (required)
├── README.md              # Documentation (recommended)
├── .env.example           # Environment template (recommended)
└── src/                   # Source files (required)
    ├── index.js
    ├── config/
    └── routes/
```

##Variable Substitution

### Simple Substitution

Use `{{VARIABLE_NAME}}` for simple string replacement:

```javascript
// Before
const projectName = '{{PROJECT_NAME}}';
const port = {{PORT}};

// After (with PROJECT_NAME=my-app, PORT=3000)
const projectName = 'my-app';
const port = 3000;
```

### Go Template Syntax

Use `{{ .VariableName }}` with helper functions:

```javascript
// Before
class {{ .ProjectName | ToPascalCase }}Service {
  constructor() {
    this.name = '{{ .ProjectName }}';
  }
}

// After (with ProjectName=my-api)
class MyApiService {
  constructor() {
    this.name = 'my-api';
  }
}
```

### Available Helper Functions

| Function | Input | Output | Description |
|----------|-------|--------|-------------|
| `ToPascalCase` | `my-project` | `MyProject` | Pascal case conversion |
| `ToCamelCase` | `my-project` | `myProject` | Camel case conversion |
| `ToUpper` | `api-key` | `API-KEY` | Uppercase conversion |

### Built-in Variables

These variables are automatically provided:

- `PROJECT_NAME`: Project name from init command
- `PORT`: Server port (default: 8080)
- `CORS_ORIGINS`: Comma-separated allowed origins

## Template Validation Rules

Templates are validated before installation:

1. **Required Files:**
   - `template.yaml` must exist
   - `src/` directory must exist

2. **Required Fields:**
   - `name` - Must be unique and URL-safe
   - `version` - Must follow semver (X.Y.Z)
   - `language` - Must be supported language

3. **Language Support:**
   - `go` - Go with Chi/Echo/Fiber
   - `nodejs-express` - Node.js with Express
   - `nodejs-fastify` - Node.js with Fastify
   - `python-fastapi` - Python with FastAPI
   - `go-graphql` - Go with gqlgen
   - `nodejs-apollo` - Node.js with Apollo Server
   - `nodejs-yoga` - Node.js with GraphQL Yoga

4. **Variable Validation:**
   - Names must be alphanumeric + underscore
   - Required variables must have no default
   - Defaults must match expected type

## Community Template Guidelines

### Best Practices

1. **Clear Documentation**
   - Include comprehensive README.md
   - Document all variables
   - Provide usage examples

2. **Minimal Dependencies**
   - Only include essential dependencies
   - Document optional dependencies
   - Pin major versions

3. **Security**
   - No hardcoded secrets
   - Use `.env.example` for configuration
   - Follow security best practices

4. **Testing**
   - Include sample tests
   - Verify template generates successfully
   - Test all variable combinations

5. **Maintenance**
   - Keep dependencies updated
   - Tag releases with semantic versions
   - Respond to issues promptly

### Template Categories

**web:** Single-page applications, APIs
**mobile:** Mobile backend services
**microservices:** Distributed systems, service mesh
**fullstack:** Full-stack applications with frontend + backend

## Template Registry

### Registry Format

The community registry is a JSON file hosted on GitHub:

```json
{
  "templates": [
    {
      "name": "ecommerce-bff",
      "description": "E-commerce backend with Stripe integration",
      "url": "https://github.com/user/ecommerce-bff",
      "author": "username",
      "tags": ["ecommerce", "stripe", "payments"],
      "language": "nodejs-express",
      "version": "2.1.0"
    }
  ],
  "last_updated": "2025-12-26T00:00:00Z"
}
```

### Submitting to Registry

1. Create template following guidelines
2. Test thoroughly
3. Tag a release on GitHub
4. Submit PR to [bffgen/templates/registry.json](https://github.com/RichGod93/bffgen/blob/main/templates/registry.json)

## Installation Process

### GitHub Installation Flow

```
1. User runs: bffgen template install user/repo
2. URL Normalization: github.com/user/repo → https://github.com/user/repo
3. Git Clone: git clone --depth 1 <url> ~/.bffgen/templates/community/repo
4. Cleanup: Remove .git directory
5. Validation: Check template.yaml and src/
6. Cache: Template ready for use
```

### Registry Installation Flow

```
1. User runs: bffgen template install ecommerce-bff
2. Registry Lookup: Find in ~/.bffgen/templates/registry.json
3. GitHub Install: Follow GitHub installation flow with URL from registry
4. Success: Template installed and cached
```

## Advanced Features

### Conditional File Generation

Skip files based on variables:

```yaml
files:
  - src/**
  - tests/** # Only if INCLUDE_TESTS=true
```

### Multi-Step Post-Install

Chain commands for complex setup:

```yaml
post_install:
  - npm install
  - npx prisma generate  # If using Prisma
  - npm run build
  - echo "✓ Setup complete"
```

### Template Inheritance (Future)

Extend base templates:

```yaml
extends: nodejs-base
name: nodejs-auth
features:
  - JWT Authentication  # Added to base features
```

## Troubleshooting

### Common Issues

**"Template not found"**
- Run `bffgen template list` to see available templates
- Verify template name spelling
- Update registry: `bffgen template registry update`

**"Validation failed"**
- Check `template.yaml` syntax
- Ensure `src/` directory exists
- Verify language is supported

**"Git clone failed"**
- Check internet connection
- Verify repository is public
- Ensure git is installed

## Examples

### Minimal Template

```yaml
name: minimal-api
version: 1.0.0
description: Minimal REST API
author: bffgen
category: web
language: nodejs-express
```

### Full-Featured Template

```yaml
name: saas-starter
version: 2.0.0
description: Production SaaS with auth, billing, admin
author: bffgen
category: fullstack
language: nodejs-express

features:
  - JWT Authentication
  - Stripe Billing
  - Admin Dashboard
  - Role-Based Access Control

variables:
  - name: DATABASE_URL
    description: PostgreSQL connection string
    required: true
  
  - name: STRIPE_SECRET_KEY
    description: Stripe API secret key
    required: true
  
  - name: ADMIN_EMAIL
    description: Initial admin email
    default: admin@example.com

files:
  - src/**
  - migrations/**
  - tests/**

post_install:
  - npm install
  - npx prisma generate
 - echo "Run 'npm run dev' to start"
```

---

**Related Documents:**
- [CONTRIBUTING_TEMPLATES.md](../CONTRIBUTING_TEMPLATES.md)
- [ARCHITECTURE_DEEP_DIVE.md](ARCHITECTURE_DEEP_DIVE.md)
- [Quick Reference](QUICK_REFERENCE.md)
