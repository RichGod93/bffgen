# BFFGen Projects

A collection of Backend-for-Frontend (BFF) example projects demonstrating the capabilities of the **bffgen** framework across different runtimes and use cases.

## Overview

This repository contains reference implementations and examples of BFF services generated using `bffgen` - a code generation tool for creating production-ready Backend-for-Frontend layers. The projects showcase real-world integrations with external APIs, best practices in BFF architecture, and implementation patterns across different programming languages.

## Repository Structure

```
bffgen-projects/
├── node/              # Node.js/Express implementations
│   └── news-aggregator/
└── go/                # Go implementations (coming soon)
```

## Projects

### Node.js Runtime

#### 📰 [News Aggregator](./node/news-aggregator/)

A production-ready BFF service that aggregates news from multiple sources (NY Times and Guardian APIs).

**Features:**

- Multi-source news aggregation (NY Times + Guardian)
- JWT authentication support
- Rate limiting and CORS configuration
- Comprehensive error handling and logging
- Production-ready security (Helmet, CORS, rate limiting)
- OpenAPI/Swagger documentation

**Quick Start:**

```bash
cd node/news-aggregator
npm install
# Add API keys to .env file
npm run dev
```

[Full Documentation →](./node/news-aggregator/NEWS_AGGREGATOR_README.md)

### Go Runtime

Go implementations are planned for future releases.

## What is BFFGen?

`bffgen` is a code generation framework that creates Backend-for-Frontend services with:

- **Multiple Runtime Support**: Node.js (Express), Go (planned)
- **Production-Ready Code**: Security, logging, error handling, monitoring
- **Best Practices**: Layered architecture (routes → controllers → services)
- **Middleware Stack**: Authentication, validation, request tracking, rate limiting
- **API Documentation**: Auto-generated OpenAPI/Swagger specs
- **Testing Setup**: Jest/testing framework integration

## Key Features Demonstrated

### Architecture Patterns

- Clean separation of concerns (routes, controllers, services)
- Centralized error handling
- Request/response transformation
- Data aggregation from multiple sources

### Production Features

- Security headers (Helmet.js)
- CORS configuration
- Rate limiting
- JWT authentication
- Structured logging (Winston)
- Request ID tracking
- Health check endpoints

### Developer Experience

- Hot reload in development
- Comprehensive error messages
- Environment-based configuration
- Testing infrastructure
- Linting and code quality tools

## Getting Started

### Prerequisites

- **Node.js**: 18+ (for Node.js projects)
- **Go**: 1.20+ (for Go projects, when available)
- **API Keys**: Depending on the project (e.g., NY Times, Guardian)

### Running a Project

1. **Choose a project directory:**

   ```bash
   cd node/news-aggregator
   ```

2. **Install dependencies:**

   ```bash
   npm install
   ```

3. **Configure environment:**

   ```bash
   # Copy .env.example to .env
   # Add your API keys and configuration
   ```

4. **Start the development server:**

   ```bash
   npm run dev
   ```

5. **Test the API:**

   ```bash
   curl http://localhost:8080/health
   ```

## Project Documentation

Each project includes detailed documentation:

- **README.md** - Quick start guide
- **PROJECT_SUMMARY.md** - Implementation details and architecture
- **API Documentation** - Swagger UI available at `/api-docs` when running

## Use Cases

These projects demonstrate BFFGen's suitability for:

- **API Aggregation**: Combining multiple backend services into a unified interface
- **Protocol Translation**: Converting between different API formats
- **Data Transformation**: Reshaping responses for frontend requirements
- **Security Layer**: Adding authentication and authorization
- **Rate Limiting**: Protecting downstream services
- **Caching**: Improving performance and reducing backend load

## Testing

Each project includes a test suite:

```bash
cd <project-directory>
npm test                 # Run all tests
npm run test:watch      # Watch mode
npm run test:coverage   # Coverage report
```

## Deployment

Projects are production-ready and include:

- Docker support (where applicable)
- Environment-based configuration
- Health check endpoints
- Graceful shutdown handling
- Process manager compatibility (PM2)

## Research Context

This repository is part of an MSc thesis exploring automated Backend-for-Frontend generation patterns, comparing different runtime implementations, and evaluating the effectiveness of code generation for BFF architectures.

## Contributing

These projects serve as reference implementations. Feel free to:

- Use them as templates for your own BFF services
- Report issues or suggest improvements
- Submit pull requests with enhancements

## License

MIT

## Resources

- **BFFGen Documentation**: [github.com/RichGod93/bffgen](https://github.com/RichGod93/bffgen)
- **BFF Pattern**: [Pattern: Backends For Frontends](https://samnewman.io/patterns/architectural/bff/)
- **Thesis Documentation**: See individual project documentation for research notes

## Support

For questions or issues:

1. Check the project-specific documentation
2. Review the BFFGen documentation
3. Open an issue on GitHub

---

**Note**: Each project is self-contained with its own dependencies, configuration, and documentation. Navigate to individual project directories for specific setup instructions.
