# {{PROJECT_NAME}}

SaaS starter project generated with bffgen.

## Features

- ✅ JWT Authentication
- ✅ Stripe Billing Integration
- ✅ Admin Dashboard
- ✅ User Management
- ✅ Rate Limiting
- ✅ Security Headers

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start development server:
```bash
npm run dev
```

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT signing
- `STRIPE_SECRET_KEY` - Stripe API key
- `PORT` - Server port (default: {{PORT}})

## API Endpoints

### Authentication
- `POST /auth/login` - User login
- `GET /auth/profile` - Get current user

### Users
- `GET /users` - List all users
- `GET /users/:id` - Get user by ID
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user

### Admin
- `GET /admin/stats` - Dashboard statistics
- `GET /admin/activity` - Recent activity

## License

MIT
