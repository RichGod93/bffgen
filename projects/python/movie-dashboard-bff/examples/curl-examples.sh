#!/bin/bash

# Movie Dashboard BFF - cURL Examples
# Demonstrates all major endpoints with example requests

BASE_URL="http://localhost:8000"
USER_SERVICE_URL="http://localhost:3001"

echo "🎬 Movie Dashboard BFF - Example Requests"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# === Health Checks ===

echo -e "${BLUE}=== Health Checks ===${NC}"
echo ""

echo -e "${GREEN}Main BFF Health:${NC}"
curl -s ${BASE_URL}/health | jq
echo ""

echo -e "${GREEN}User Service Health:${NC}"
curl -s ${USER_SERVICE_URL}/health | jq
echo ""
echo ""

# === Dashboard Endpoints (Aggregated) ===

echo -e "${BLUE}=== Dashboard Endpoints (BFF Aggregation) ===${NC}"
echo ""

echo -e "${GREEN}1. Get Personalized Feed (Popular + User Data):${NC}"
curl -s "${BASE_URL}/api/dashboard/feed?page=1" | jq '.movies[0:3] | .[].title'
echo ""

echo -e "${GREEN}2. Get Complete Dashboard (Multiple Sources):${NC}"
curl -s "${BASE_URL}/api/dashboard/complete" | jq '{
  popular: .popular_movies | length,
  trending: .trending_movies | length,
  stats: .stats
}'
echo ""

echo -e "${GREEN}3. Search Movies (Enriched with User Data):${NC}"
curl -s "${BASE_URL}/api/dashboard/search/enriched?query=inception" | jq '.results[0] | {
  title,
  is_favorite,
  vote_average
}'
echo ""

echo -e "${GREEN}4. Get Enriched Movie Details (TMDB + User Data):${NC}"
curl -s "${BASE_URL}/api/dashboard/movie/550/enriched" | jq '{
  title,
  vote_average,
  user_data,
  cast: .cast[0:3] | .[].name
}'
echo ""
echo ""

# === Direct TMDB Endpoints ===

echo -e "${BLUE}=== Direct TMDB Endpoints ===${NC}"
echo ""

echo -e "${GREEN}5. Get Popular Movies:${NC}"
curl -s "${BASE_URL}/api/movies/popular?page=1" | jq '.results[0:3] | .[].title'
echo ""

echo -e "${GREEN}6. Get Movie Details:${NC}"
curl -s "${BASE_URL}/api/movies/550" | jq '{title, tagline, vote_average}'
echo ""

echo -e "${GREEN}7. Search Movies:${NC}"
curl -s "${BASE_URL}/api/movies/search?query=fight+club&page=1" | jq '.results[0] | {title, release_date}'
echo ""
echo ""

# === User Service Endpoints ===

echo -e "${BLUE}=== User Service Endpoints ===${NC}"
echo ""

echo -e "${GREEN}8. Get User Favorites:${NC}"
curl -s "${USER_SERVICE_URL}/favorites" | jq '.favorites | .[].title'
echo ""

echo -e "${GREEN}9. Add to Favorites:${NC}"
curl -s -X POST "${USER_SERVICE_URL}/favorites" \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 278,
    "title": "The Shawshank Redemption",
    "rating": 5,
    "poster_path": "/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg"
  }' | jq '{message, favorite: .favorite.title}'
echo ""

echo -e "${GREEN}10. Get Watchlist:${NC}"
curl -s "${USER_SERVICE_URL}/watchlist" | jq '.watchlist | .[].title'
echo ""

echo -e "${GREEN}11. Add to Watchlist:${NC}"
curl -s -X POST "${USER_SERVICE_URL}/watchlist" \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 27205,
    "title": "Inception",
    "release_date": "2010-07-16"
  }' | jq '{message, item: .watchlist_item.title}'
echo ""

echo -e "${GREEN}12. Delete from Favorites:${NC}"
curl -s -X DELETE "${USER_SERVICE_URL}/favorites/278" | jq
echo ""
echo ""

# === Advanced Examples ===

echo -e "${BLUE}=== Advanced Examples ===${NC}"
echo ""

echo -e "${GREEN}13. Pagination Example:${NC}"
curl -s "${BASE_URL}/api/dashboard/feed?page=2" | jq '{
  current_page,
  total_pages,
  movie_count: .movies | length
}'
echo ""

echo -e "${GREEN}14. Full Movie Details with Everything:${NC}"
curl -s "${BASE_URL}/api/dashboard/movie/155/enriched" | jq '{
  title,
  runtime,
  genres: .genres | .[].name,
  cast_count: .cast | length,
  user_data
}'
echo ""

echo -e "${GREEN}15. Reset User Service Data (Testing):${NC}"
curl -s -X POST "${USER_SERVICE_URL}/reset" | jq
echo ""
echo ""

# === OpenAPI Documentation ===

echo -e "${BLUE}=== Documentation ===${NC}"
echo ""
echo "📖 Interactive API Docs:"
echo "   Main BFF:     ${BASE_URL}/docs"
echo "   User Service: ${USER_SERVICE_URL}/docs"
echo ""
echo "📊 OpenAPI Schema:"
echo "   ${BASE_URL}/openapi.json"
echo ""
echo ""

echo "✅ All examples completed!"
echo ""
echo "💡 Tips:"
echo "   - Use jq for pretty JSON formatting"
echo "   - Add --verbose or -v for request details"
echo "   - Use -i to see response headers"
echo "   - Chain commands with && or ||"
echo ""

