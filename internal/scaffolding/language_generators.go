package scaffolding

import (
	"fmt"
	"path/filepath"
)

// LanguageType represents the programming language/runtime
type LanguageType string

const (
	LanguageGo            LanguageType = "go"
	LanguageNodeExpress   LanguageType = "nodejs-express"
	LanguageNodeFastify   LanguageType = "nodejs-fastify"
	LanguagePythonFastAPI LanguageType = "python-fastapi"
	LanguageNodeApollo    LanguageType = "nodejs-apollo" // GraphQL Apollo Server
	LanguageNodeYoga      LanguageType = "nodejs-yoga"   // GraphQL Yoga
	LanguageGoGqlgen      LanguageType = "go-graphql"    // GraphQL gqlgen
)

// LanguageConfig holds configuration for different languages
type LanguageConfig struct {
	Type         LanguageType
	Name         string
	Framework    string
	DepsPath     string
	MainFile     string
	TemplatePath string
}

// GetLanguageConfig returns configuration for a language type
func GetLanguageConfig(langType LanguageType) LanguageConfig {
	switch langType {
	case LanguageGo:
		return LanguageConfig{
			Type:         LanguageGo,
			Name:         "Go (Chi)",
			Framework:    "chi", // default framework
			DepsPath:     "go.mod",
			MainFile:     "main.go",
			TemplatePath: "templates/go",
		}
	case LanguageNodeExpress:
		return LanguageConfig{
			Type:         LanguageNodeExpress,
			Name:         "Node.js (Express)",
			Framework:    "express",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "templates/nodejs-express",
		}
	case LanguageNodeFastify:
		return LanguageConfig{
			Type:         LanguageNodeFastify,
			Name:         "Node.js (Fastify)",
			Framework:    "fastify",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "templates/nodejs-fastify",
		}
	case LanguagePythonFastAPI:
		return LanguageConfig{
			Type:         LanguagePythonFastAPI,
			Name:         "Python (FastAPI)",
			Framework:    "fastapi",
			DepsPath:     "requirements.txt",
			MainFile:     "main.py",
			TemplatePath: "templates/python-fastapi",
		}
	case LanguageNodeApollo:
		return LanguageConfig{
			Type:         LanguageNodeApollo,
			Name:         "Node.js (Apollo GraphQL)",
			Framework:    "apollo",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "templates/nodejs-apollo",
		}
	case LanguageNodeYoga:
		return LanguageConfig{
			Type:         LanguageNodeYoga,
			Name:         "Node.js (GraphQL Yoga)",
			Framework:    "yoga",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "templates/nodejs-yoga",
		}
	case LanguageGoGqlgen:
		return LanguageConfig{
			Type:         LanguageGoGqlgen,
			Name:         "Go (gqlgen GraphQL)",
			Framework:    "gqlgen",
			DepsPath:     "go.mod",
			MainFile:     "server.go",
			TemplatePath: "templates/go-graphql",
		}
	default:
		return LanguageConfig{
			Type:         LanguageGo,
			Name:         "Go",
			Framework:    "chi",
			DepsPath:     "go.mod",
			MainFile:     "main.go",
			TemplatePath: "templates/go",
		}
	}
}

// GetSupportedLanguages returns all supported language types
func GetSupportedLanguages() []LanguageConfig {
	return []LanguageConfig{
		GetLanguageConfig(LanguageGo),
		GetLanguageConfig(LanguageNodeExpress),
		GetLanguageConfig(LanguageNodeFastify),
		GetLanguageConfig(LanguagePythonFastAPI),
		GetLanguageConfig(LanguageNodeApollo),
		GetLanguageConfig(LanguageNodeYoga),
		GetLanguageConfig(LanguageGoGqlgen),
	}
}

// IsValidLanguage checks if a language type is supported
func IsValidLanguage(lang string) bool {
	langType := LanguageType(lang)
	for _, config := range GetSupportedLanguages() {
		if string(config.Type) == string(langType) {
			return true
		}
	}
	return false
}

// GetLanguageFromFramework maps framework names to language types
func GetLanguageFromFramework(framework string) LanguageType {
	switch framework {
	case "chi", "echo", "fiber":
		return LanguageGo
	case "express":
		return LanguageNodeExpress
	case "fastify":
		return LanguageNodeFastify
	case "apollo":
		return LanguageNodeApollo
	case "yoga":
		return LanguageNodeYoga
	case "gqlgen":
		return LanguageGoGqlgen
	case "python-fastapi", "fastapi":
		return LanguagePythonFastAPI
	default:
		return LanguageGo
	}
}

// GenerateProjectFile creates project files based on language type
func (g *Generator) GenerateProjectFile(projectName string, langConfig LanguageConfig) error {
	switch langConfig.Type {
	case LanguageGo:
		return g.generateGoProject(projectName, langConfig)
	case LanguageNodeExpress:
		return g.generateNodeExpressProject(projectName, langConfig)
	case LanguageNodeFastify:
		return g.generateNodeFastifyProject(projectName, langConfig)
	case LanguagePythonFastAPI:
		return g.generatePythonFastAPIProject(projectName, langConfig)
	default:
		return fmt.Errorf("unsupported language type: %s", langConfig.Type)
	}
}

// generateGoProject creates Go-specific project structure
func (g *Generator) generateGoProject(projectName string, config LanguageConfig) error {
	// This will contain the Go project generation logic
	// Implementation will be moved from init.go
	return nil
}

// generateNodeExpressProject creates Node.js Express project structure
func (g *Generator) generateNodeExpressProject(projectName string, config LanguageConfig) error {
	// Create package.json
	packageJson := g.generatePackageContent(projectName, "express")
	return g.generateFile(filepath.Join(projectName, "package.json"), packageJson)
}

// generateNodeFastifyProject creates Node.js Fastify project structure
func (g *Generator) generateNodeFastifyProject(projectName string, config LanguageConfig) error {
	// Create package.json
	packageJson := g.generatePackageContent(projectName, "fastify")
	return g.generateFile(filepath.Join(projectName, "package.json"), packageJson)
}

// generatePythonFastAPIProject creates Python FastAPI project structure
func (g *Generator) generatePythonFastAPIProject(projectName string, config LanguageConfig) error {
	// Create requirements.txt
	requirements := g.generatePythonRequirements(projectName)
	return g.generateFile(filepath.Join(projectName, "requirements.txt"), requirements)
}

// generatePythonRequirements creates requirements.txt content for Python projects
func (g *Generator) generatePythonRequirements(projectName string) string {
	return `fastapi==0.109.0
uvicorn[standard]==0.27.0
pydantic==2.5.3
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
python-multipart==0.0.6
slowapi==0.1.9
httpx==0.26.0
redis==5.0.1
`
}

// generatePackageContent creates package.json content for Node.js projects
func (g *Generator) generatePackageContent(projectName, framework string) string {
	var dependencies string
	var mainContent string

	switch framework {
	case "express":
		dependencies = `    "express": "^4.18.2",
    "cors": "^2.8.5",
    "helmet": "^7.1.0",
    "express-rate-limit": "^7.1.5",
    "jsonwebtoken": "^9.0.2",
    "cookie-parser": "^1.4.6",
    "express-validator": "^7.0.1"`
		mainContent = g.generateExpressMainContent(projectName)
	case "fastify":
		dependencies = `    "@fastify/cors": "^8.3.0",
    "@fastify/helmet": "^11.1.1",
    "@fastify/rate-limit": "^7.7.0",
    "fastify": "^4.24.3",
    "fastify-jwt": "^7.2.0",
    "fastify-cookie": "^8.3.0"`
		mainContent = g.generateFastifyMainContent(projectName)
	default:
		return ""
	}

	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Backend-for-Frontend service generated by bffgen",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js",
    "test": "jest"
  },
  "keywords": ["bff", "backend-for-frontend", "api"],
  "author": "",
  "license": "MIT",
  "dependencies": {
%s
  },
  "devDependencies": {
    "nodemon": "^3.0.1",
    "jest": "^29.7.0"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}`, projectName, dependencies) + "\n" + mainContent
}

// generateExpressMainContent creates Express.js main server content
func (g *Generator) generateExpressMainContent(projectName string) string {
	return `
// Generated Express.js BFF server
const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
const cookieParser = require('cookie-parser');
const { body, validationResult } = require('express-validator');

const app = express();
const PORT = process.env.PORT || 8080;

// Middleware
app.use(helmet());
app.use(cookieParser());
app.use(express.json({ limit: '5mb' }));
app.use(express.urlencoded({ extended: true }));

// Rate limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // limit each IP to 100 requests per windowMs
  message: 'Too many requests from this IP, please try again later.'
});
app.use(limiter);

// CORS configuration
app.use(cors({
  origin: process.env.CORS_ORIGINS?.split(',') || ['http://localhost:3000'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Accept', 'Authorization', 'Content-Type', 'X-CSRF-Token']
}));

// Security headers middleware
app.use((req, res, next) => {
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-Frame-Options', 'DENY');
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
  next();
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// Auth endpoints placeholder
app.get('/api/auth/profile', (req, res) => {
  res.json({ message: 'Auth endpoint - implement your authentication logic' });
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Something went wrong!' });
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({ error: 'Route not found' });
});

// Start server
app.listen(PORT, () => {
  console.log("🚀 BFF server starting on port " + PORT);
});

module.exports = app;
`
}

// generateFastifyMainContent creates Fastify main server content
func (g *Generator) generateFastifyMainContent(projectName string) string {
	return `
// Generated Fastify BFF server
const fastify = require('fastify')({ 
  logger: {
    level: 'info',
    prettyPrint: process.env.NODE_ENV === 'development' ? true : false
  }
});

const PORT = process.env.PORT || 8080;

// Register plugins
async function start() {
  // CORS configuration
  await fastify.register(require('@fastify/cors'), {
    origin: process.env.CORS_ORIGINS?.split(',') || ['http://localhost:3000'],
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Accept', 'Authorization', 'Content-Type', 'X-CSRF-Token']
  });

  // Security headers
  await fastify.register(require('@fastify/helmet'), {
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"],
        styleSrc: ["'self'", "'unsafe-inline'"],
        scriptSrc: ["'self'"],
        imgSrc: ["'self'", "data:"],
        fontSrc: ["'self'"],
        connectSrc: ["'self'"],
        frameAncestors: ["'none'"]
      }
    }
  });

  // Rate limiting
  await fastify.register(require('@fastify/rate-limit'), {
    max: 100,
    timeWindow: '15 minutes'
  });

  // Cookie support
  await fastify.register(require('fastify-cookie'));

  // Health check endpoint
  fastify.get('/health', async (request, reply) => {
    return { status: 'healthy', timestamp: new Date().toISOString() };
  });

  // Auth endpoints placeholder
  fastify.get('/api/auth/profile', async (request, reply) => {
    return { message: 'Auth endpoint - implement your authentication logic' };
  });

  // Error handler
  fastify.setErrorHandler((error, request, reply) => {
    fastify.log.error(error);
    reply.status(500).send({ error: 'Something went wrong!' });
  });

  // 404 handler
  fastify.setNotFoundHandler((request, reply) => {
    reply.status(404).send({ error: 'Route not found' });
  });

  // Start server
  try {
    await fastify.listen({ port: PORT, host: '0.0.0.0' });
    fastify.log.info('🚀 BFF server starting on port ' + PORT);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
}

start();
`
}

// generateFile creates a file with given content
func (g *Generator) generateFile(filepath, content string) error {
	// Implementation will handle file creation
	return nil
}
