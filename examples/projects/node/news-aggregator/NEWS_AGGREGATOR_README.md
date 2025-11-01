# News API Aggregator

A Backend-for-Frontend (BFF) service that aggregates news from multiple sources: **NY Times** and **Guardian**.

## Features

- 🌐 Aggregates news from multiple APIs
- 🔒 JWT authentication support
- 🚀 Production-ready with security headers
- 📊 Rate limiting and CORS configured
- 📝 Comprehensive logging
- ⚡ Built with Express.js using bffgen

## Prerequisites

- Node.js 18+
- API Keys:
  - [NY Times API Key](https://developer.nytimes.com/)
  - [Guardian API Key](https://open-platform.theguardian.com/access/)

## Installation

1. **Install dependencies:**

   ```bash
   npm install
   ```

2. **Configure API keys:**
   Edit `.env` file and add your API keys:

   ```env
   NYTIMES_API_KEY=your-nytimes-api-key-here
   GUARDIAN_API_KEY=your-guardian-api-key-here
   ```

3. **Start the server:**
   ```bash
   npm run dev
   ```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check

```bash
GET /health
```

Returns server health status.

### NY Times Endpoints

#### Get Headlines

```bash
GET /api/news/nytimes/headlines
```

**Query Parameters:**

- `section` (optional): Section name (e.g., world, business, technology)

**Example:**

```bash
curl http://localhost:8080/api/news/nytimes/headlines
```

#### Search Articles

```bash
GET /api/news/nytimes/search?q=your-search-term
```

**Query Parameters:**

- `q` (required): Search query
- `page` (optional): Page number
- `sort` (optional): Sort order (newest, oldest, relevance)

**Example:**

```bash
curl "http://localhost:8080/api/news/nytimes/search?q=climate+change"
```

### Guardian Endpoints

#### Get Headlines

```bash
GET /api/news/guardian/headlines
```

**Query Parameters:**

- `section` (optional): Section name (e.g., world, business, technology)
- `page-size` (optional): Number of results (default: 10)

**Example:**

```bash
curl "http://localhost:8080/api/news/guardian/headlines?section=world&page-size=20"
```

#### Search Articles

```bash
GET /api/news/guardian/search?q=your-search-term
```

**Query Parameters:**

- `q` (required): Search query
- `page` (optional): Page number
- `page-size` (optional): Number of results per page
- `order-by` (optional): Sort order (newest, oldest, relevance)

**Example:**

```bash
curl "http://localhost:8080/api/news/guardian/search?q=technology&page-size=15"
```

## Configuration

All configuration is in `.env` file:

```env
# Server
PORT=8080
NODE_ENV=development

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:5173

# NY Times
NYTIMES_API_KEY=your-key-here
NYTIMES_TIMEOUT=10000
NYTIMES_RETRIES=2

# Guardian
GUARDIAN_API_KEY=your-key-here
GUARDIAN_TIMEOUT=10000
GUARDIAN_RETRIES=2
```

## Development

### Available Scripts

- `npm start` - Start production server
- `npm run dev` - Start development server with hot reload
- `npm test` - Run tests
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix linting issues

### Project Structure

```
news-aggregator/
├── src/
│   ├── index.js              # Main server file
│   ├── routes/               # Route handlers
│   │   ├── nytimes.js       # NY Times routes
│   │   └── guardian.js      # Guardian routes
│   ├── controllers/          # Business logic (optional)
│   ├── services/            # HTTP clients
│   ├── middleware/          # Auth, logging, validation
│   └── utils/               # Helper functions
├── tests/                   # Test files
├── .env                     # Environment variables (create from .env.example)
├── bffgen.config.json       # BFF configuration
└── package.json
```

## Security

- **Rate Limiting:** 100 requests per 15 minutes per IP
- **CORS:** Configured for specific origins
- **Helmet:** Security headers enabled
- **JWT:** Authentication support (optional)

## Error Handling

All endpoints return standardized error responses:

```json
{
  "error": "Service name",
  "message": "Error description",
  "details": "Additional details (dev only)"
}
```

## Testing

Test the endpoints:

```bash
# Health check
curl http://localhost:8080/health

# NY Times headlines
curl http://localhost:8080/api/news/nytimes/headlines

# Guardian search
curl "http://localhost:8080/api/news/guardian/search?q=sports"
```

## API Rate Limits

- **NY Times:** 500 requests per day, 5 requests per minute
- **Guardian:** 5,000 requests per day

## Troubleshooting

### Server won't start

- Check Node.js version: `node --version` (should be 18+)
- Verify API keys are set in `.env`
- Check port 8080 is not in use

### API returns 401/403

- Verify your API keys are valid
- Check if you've exceeded rate limits

### No data returned

- Check your internet connection
- Verify the API services are online
- Check the query parameters

## Production Deployment

1. Set `NODE_ENV=production` in `.env`
2. Use a process manager like PM2:
   ```bash
   npm install -g pm2
   pm2 start src/index.js --name news-aggregator
   ```
3. Configure nginx as reverse proxy
4. Enable HTTPS
5. Set strong JWT_SECRET

## License

MIT

## Support

For issues and questions:

- Check the [bffgen documentation](https://github.com/RichGod93/bffgen)
- Open an issue on GitHub
