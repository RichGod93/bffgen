# Contributing Templates to BFFGen

We welcome community contributions to the BFFGen Template Gallery! This guide will help you create high-quality templates that can be used by developers worldwide.

## Template Structure

A valid BFFGen template must have the following structure:

```
my-template-name/
├── template.yaml        # Manifest file (Required)
├── src/                 # Source code (Required)
├── config/              # Configuration files
├── tests/               # Tests
├── .env.example         # Environment variables template
└── README.md            # Template documentation
```

## The Manifest (template.yaml)

The `template.yaml` file is the heart of your template. It defines metadata, features, and variables.

### Example

```yaml
name: my-awesome-template
version: 1.0.0
description: A description of what this template does
author: your-github-username
category: mobile  # web, mobile, microservices, various
language: nodejs-express # nodejs-express, nodejs-fastify, go, python-fastapi

features:
  - "GraphQL API"
  - "Redis Caching"

variables:
  - name: PROJECT_NAME
    description: Name of the project
    default: my-project
  
  - name: API_KEY
    description: External API Key
    required: true

post_install:
  - "npm install"
  - "echo 'Don't forget to set your API_KEY!'"
```

## Best Practices

1.  **Clean Architecture**: Follow standard patterns for the chosen language/framework.
2.  **Documentation**: Include a clear `README.md` explaining how to run and deploy the project.
3.  **Environment Variables**: Use `.env.example` for all configurable values. Never commit secrets.
4.  **Tests**: Include basic tests (unit/integration) to ensure the generated project works out of the box.
5.  **Linting**: Ensure code follows standard style guides (ESLint, Go fmt, Black/Pylint).

## Variable Substitution

Files in your template can use Go template syntax `{{ .VariableName }}` to substitute values provided during initialization.

Supported default variables:
*   `{{ .PROJECT_NAME }}`

## Submission Process

1.  Host your template in a public GitHub repository.
2.  Validate your template locally ensuring `bffgen init --template path/to/local/template` works.
3.  Submit a Pull Request to [BFFGen Repository](https://github.com/RichGod93/bffgen) to add your template to the Community Registry `registry.json`.

## Testing Your Template

To test your template locally:

1.  Create your template directory.
2.  Run `bffgen init my-test-project --template ./path/to/your/template`.
3.  Verify the project is created correctly and runs.
