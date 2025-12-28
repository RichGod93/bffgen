# GraphQL Support in bffgen

bffgen now supports generating production-ready GraphQL BFF (Backend-for-Frontend) services with comprehensive features including REST-to-GraphQL aggregation, schema stitching, and automatic type generation.

## Table of Contents

- [Overview](#overview)
- [Supported GraphQL Frameworks](#supported-graphql-frameworks)
- [Quick Start](#quick-start)
- [Features](#features)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [REST-to-GraphQL Aggregation](#rest-to-graphql-aggregation)
- [Schema Stitching](#schema-stitching)
- [Type Generation](#type-generation)
- [Best Practices](#best-practices)

## Overview

GraphQL BFF services act as an aggregation layer between your frontend applications and backend microservices. They provide:

- **Unified API**: Single GraphQL endpoint for all your frontend data needs
- **Data Aggregation**: Combine data from multiple microservices in a single query
- **Type Safety**: Auto-generated TypeScript/Go types from your GraphQL schema
- **Performance**: Built-in caching, batching, and query optimization
- **Flexibility**: REST-to-GraphQL bridging for legacy services

## Supported GraphQL Frameworks

### Node.js Frameworks

1. **Apollo Server 4** (`nodejs-apollo`)
   - Industry-standard GraphQL server
   - Rich plugin ecosystem
   - Advanced features like Apollo Federation
   - Best for: Large-scale applications, federation needs

2. **GraphQL Yoga** (`nodejs-yoga`)
   - Lightweight and modern
   - Built-in subscriptions
   - Better performance than Apollo
   - Best for: New projects, performance-critical applications

### Go Framework

3. **gqlgen** (`go-graphql`)
   - Type-safe Go GraphQL library
   - Schema-first approach
   - High performance
   - Best for: Go microservices, performance-critical systems

## Quick Start

### 1. Initialize a GraphQL BFF Project

**Using Apollo Server:**
```bash
bffgen init my-graphql-bff --lang nodejs-apollo
cd my-graphql-bff
npm install
npm run dev
```

**Using GraphQL Yoga:**
```bash
bffgen init my-graphql-bff --lang nodejs-yoga
cd my-graphql-bff
npm install
npm run dev
```

**Using Go gqlgen:**
```bash
bffgen init my-graphql-bff --lang go-graphql
cd my-graphql-bff
go mod download
go run .
```

**Using Template:**
```bash
bffgen init my-graphql-bff --template graphql-api
cd my-graphql-bff
npm install
npm run dev
```

### 2. Access GraphQL Playground

Once your server is running, open your browser to:
- **Apollo Server**: http://localhost:4000/graphql
- **GraphQL Yoga**: http://localhost:4000/graphql
- **gqlgen**: http://localhost:8080/graphql

## Features

### ✨ Core Features

1. **REST-to-GraphQL Aggregation**
   - Wrap REST APIs as GraphQL data sources
   - Automatic request/response transformation
   - Field-level data fetching

2. **Schema Stitching**
   - Combine multiple GraphQL schemas
   - Remote schema delegation
   - Type conflict resolution

3. **Type Generation**
   - TypeScript types from GraphQL schema
   - Go types from GraphQL schema
   - Watch mode for development

4. **Authentication & Authorization**
   - JWT token validation
   - Role-based access control
   - Custom authentication plugins

5. **Performance Optimizations**
   - Response caching with TTL
   - Request batching with DataLoader
   - Query depth limiting
   - Circuit breaker pattern

6. **Developer Experience**
   - GraphQL Playground / GraphiQL
   - Hot reloading in development
   - Comprehensive error handling
   - Request tracing and logging

## Usage Examples

### Example 1: User Dashboard Aggregation

This example shows how to aggregate data from multiple microservices into a single GraphQL query:

**GraphQL Query:**
```graphql
query UserDashboard($userId: ID!) {
  dashboard(userId: $userId) {
    user {
      id
      name
      email
    }
    recentActivities {
      id
      type
      description
      timestamp
    }
    notifications {
      id
      title
      read
    }
    stats {
      totalOrders
      totalSpent
      loyaltyPoints
    }
  }
}
```

**Generated Resolver** (automatically aggregates from 4 microservices):
```javascript
userDashboard: async (_, { userId }, { dataSources }) => {
  // Parallel requests to multiple services
  const [user, activities, notifications, stats] = await Promise.all([
    dataSources.userService.getUserById(userId),
    dataSources.activityService.getRecentActivity(userId),
    dataSources.notificationService.getNotifications(userId),
    dataSources.analyticsService.getUserStats(userId),
  ]);

  return { user, recentActivity: activities, notifications, stats };
}
```

### Example 2: Schema Stitching

Combine multiple remote GraphQL schemas:

**Configuration:**
```javascript
const { stitchRemoteSchemas } = require('./utils/schema-stitching');

const remoteSchemas = [
  {
    name: 'users',
    url: 'http://localhost:8001/graphql',
    headers: { 'x-api-key': process.env.USERS_API_KEY },
  },
  {
    name: 'products',
    url: 'http://localhost:8002/graphql',
    headers: { 'x-api-key': process.env.PRODUCTS_API_KEY },
  },
];

const stitchedSchema = await stitchRemoteSchemas(remoteSchemas, localSchema);
```

### Example 3: Type Generation

Generate TypeScript types from your GraphQL schema:

**Run Code Generator:**
```bash
npm run codegen
```

**Generated Types** (`generated/types.ts`):
```typescript
export type User = {
  __typename?: 'User';
  id: Scalars['ID'];
  email: Scalars['String'];
  name?: Maybe<Scalars['String']>;
  createdAt: Scalars['DateTime'];
};

export type UserDashboard = {
  __typename?: 'UserDashboard';
  user: User;
  recentActivity: Array<Activity>;
  notifications: Array<Notification>;
  stats: UserStats;
};
```

## Configuration

### Project Configuration

**Apollo Server** (`package.json`):
```json
{
  "name": "my-graphql-bff",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js",
    "codegen": "graphql-codegen --config codegen.yml",
    "test": "jest --coverage"
  },
  "dependencies": {
    "@apollo/server": "^4.9.5",
    "@apollo/datasource-rest": "^6.2.2",
    "graphql": "^16.8.1"
  }
}
```

**GraphQL Yoga** (`package.json`):
```json
{
  "dependencies": {
    "graphql": "^16.8.1",
    "graphql-yoga": "^5.1.0",
    "@graphql-yoga/plugin-response-cache": "^3.2.0"
  }
}
```

### Environment Variables

Create a `.env` file:

```bash
# Server
PORT=4000
NODE_ENV=development

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:3001

# Authentication
JWT_SECRET=your-secret-key
TOKEN_EXPIRATION=7d

# Backend Services
USER_SERVICE_URL=http://localhost:8001
PRODUCT_SERVICE_URL=http://localhost:8002
ORDER_SERVICE_URL=http://localhost:8003

# Cache
CACHE_TTL=60
REDIS_URL=redis://localhost:6379

# GraphQL
GRAPHQL_DEPTH_LIMIT=10
GRAPHQL_COMPLEXITY_LIMIT=1000
```

## REST-to-GraphQL Aggregation

### Data Source Implementation

The generated BFF includes enhanced REST data sources with built-in features:

**Features:**
- ✅ Automatic request batching
- ✅ Response caching with TTL
- ✅ Circuit breaker pattern
- ✅ Retry logic with exponential backoff
- ✅ Request/response transformation
- ✅ Authentication header forwarding

**Example Data Source:**
```javascript
class UserService extends EnhancedRESTDataSource {
  constructor(options) {
    super({ ...options, serviceName: 'users' });
    this.baseURL = process.env.USER_SERVICE_URL;
  }

  async getUserById(id) {
    return this.getWithRetry(`/users/${id}`, {
      cacheOptions: { ttl: 60 }, // Cache for 60 seconds
    });
  }

  async batchGetUsers(ids) {
    return this.batchGet(ids, '/users/:id');
  }
}
```

### Mapping REST to GraphQL Types

**GraphQL Schema:**
```graphql
type User {
  id: ID!
  email: String!
  name: String
  createdAt: DateTime!
}
```

**REST API Response:**
```json
{
  "user_id": "123",
  "email_address": "user@example.com",
  "full_name": "John Doe",
  "created_timestamp": "2024-01-01T00:00:00Z"
}
```

**Transformation:**
```javascript
async getUserById(id) {
  const data = await this.get(`/users/${id}`);

  return this.transformResponse(data, {
    user_id: 'id',
    email_address: 'email',
    full_name: 'name',
    created_timestamp: 'createdAt',
  });
}
```

## Schema Stitching

### Basic Schema Stitching

Combine multiple remote GraphQL schemas into one unified schema:

```javascript
const { stitchRemoteSchemas } = require('./utils/schema-stitching');

const remoteSchemas = [
  {
    name: 'users',
    url: 'http://localhost:8001/graphql',
    batch: true,
    required: true,
  },
  {
    name: 'products',
    url: 'http://localhost:8002/graphql',
    batch: true,
    required: true,
  },
];

const stitchedSchema = await stitchRemoteSchemas(remoteSchemas);
```

### Handling Type Conflicts

When stitching schemas with conflicting types, use namespacing:

```javascript
const config = {
  name: 'products',
  url: 'http://localhost:8002/graphql',
  namespace: 'Products', // Prefix types with "Products_"
};
```

## Type Generation

### TypeScript Type Generation

**Configure GraphQL Code Generator** (`codegen.yml`):
```yaml
schema: "./schema.graphql"
generates:
  ./generated/types.ts:
    plugins:
      - "typescript"
      - "typescript-resolvers"
    config:
      useIndexSignature: true
      scalars:
        DateTime: Date
        JSON: any
```

**Run Generator:**
```bash
npm run codegen        # Generate once
npm run codegen:watch  # Watch mode
```

**Use Generated Types:**
```typescript
import { User, Resolvers } from './generated/types';

const resolvers: Resolvers = {
  Query: {
    me: async (_, __, context): Promise<User> => {
      return context.dataSources.userService.getUserById(context.user.id);
    },
  },
};
```

### Go Type Generation

For Go projects using gqlgen, types are generated automatically from your schema:

```bash
go run github.com/99designs/gqlgen generate
```

## Best Practices

### 1. Query Depth Limiting

Protect against malicious deeply nested queries:

```javascript
// Apollo Server
plugins: [
  useDepthLimit({ maxDepth: 10 })
]

// GraphQL Yoga
useDepthLimit({ maxDepth: 10 })
```

### 2. Response Caching

Cache expensive queries to reduce backend load:

```javascript
// Per-resolver caching
getUserById(id) {
  return this.get(`/users/${id}`, {
    cacheOptions: { ttl: 60 }, // 60 seconds
  });
}
```

### 3. Request Batching

Use DataLoader to batch requests:

```javascript
const userLoader = new DataLoader(async (ids) => {
  return await dataSources.userService.batchGetUsers(ids);
});

// In resolver
const user = await context.loaders.user.load(userId);
```

### 4. Error Handling

Provide meaningful errors to clients:

```javascript
if (!context.user) {
  throw new GraphQLError('Authentication required', {
    extensions: { code: 'UNAUTHENTICATED' },
  });
}
```

### 5. Monitoring and Logging

Log all GraphQL operations:

```javascript
plugins: [
  {
    async requestDidStart() {
      return {
        async didEncounterErrors(requestContext) {
          logger.error('GraphQL Error:', requestContext.errors);
        },
      };
    },
  },
]
```

## Performance Tips

1. **Enable Caching**: Cache responses at the data source level
2. **Use Batching**: Batch requests to reduce network overhead
3. **Query Depth Limiting**: Prevent abuse with depth limits
4. **Field-Level Caching**: Cache individual fields when possible
5. **Circuit Breakers**: Fail fast when services are down
6. **Connection Pooling**: Reuse HTTP connections

## Testing

### Testing GraphQL Queries

```bash
# Using curl
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ health { status } }"}'

# Using GraphQL Playground
# Open http://localhost:4000/graphql in your browser
```

### Integration Tests

```javascript
const { ApolloServer } = require('@apollo/server');
const { typeDefs, resolvers } = require('./schema');

describe('GraphQL API', () => {
  let server;

  beforeAll(async () => {
    server = new ApolloServer({ typeDefs, resolvers });
  });

  it('should return health status', async () => {
    const result = await server.executeOperation({
      query: '{ health { status } }',
    });

    expect(result.data.health.status).toBe('healthy');
  });
});
```

## Troubleshooting

### Common Issues

1. **"Circuit breaker is open"**
   - Backend service is down or timing out
   - Check service health and network connectivity

2. **"Authentication required"**
   - Missing or invalid JWT token
   - Include `Authorization: Bearer <token>` header

3. **"Query depth exceeded"**
   - Query is too deeply nested
   - Simplify your query or increase depth limit

4. **Slow queries**
   - Enable caching for expensive operations
   - Use batching with DataLoader
   - Add indexes to backend databases

## Next Steps

- [ ] Add custom scalars for your domain types
- [ ] Implement subscriptions for real-time updates
- [ ] Set up Apollo Federation for multi-team development
- [ ] Add query complexity analysis
- [ ] Implement field-level authentication
- [ ] Set up monitoring with Apollo Studio or similar

## Resources

- [Apollo Server Documentation](https://www.apollographql.com/docs/apollo-server/)
- [GraphQL Yoga Documentation](https://the-guild.dev/graphql/yoga-server)
- [gqlgen Documentation](https://gqlgen.com/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [bffgen GitHub Repository](https://github.com/RichGod93/bffgen)

## Contributing

Found a bug or have a feature request? Please open an issue on our [GitHub repository](https://github.com/RichGod93/bffgen/issues).

## License

This feature is part of bffgen, which is MIT licensed.
