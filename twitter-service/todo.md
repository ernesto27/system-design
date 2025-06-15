# Social Media Feed System Architecture Plan (Twitter-like)

## **Phase 1: Core System Design**

### **1.1 Requirements Analysis**
**Functional Requirements:**
- User registration/authentication
- Post creation (text, images, videos)
- Follow/unfollow users
- Timeline generation (home feed)
- Like, retweet, comment functionality
- User profile management

**Non-Functional Requirements:**
- Handle 100M+ daily active users
- Sub-second feed generation
- 99.9% availability
- Global content delivery

### **1.2 High-Level Architecture**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   API Gateway   │    │  Load Balancer  │
│ (Web/Mobile)    │◄──►│                 │◄──►│                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │ User Service │ │ Post Service│ │Feed Service│
        └──────────────┘ └─────────────┘ └────────────┘
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │ User DB      │ │ Post DB     │ │ Cache      │
        │ (PostgreSQL) │ │ (Cassandra) │ │ (Redis)    │
        └──────────────┘ └─────────────┘ └────────────┘
```

## **Phase 2: Database Design**

### **2.1 Core Entities**

**Users Table:**
```sql
- user_id (UUID, Primary Key)
- username (String, Unique)
- email (String, Unique)  
- bio (Text)
- follower_count (Integer)
- following_count (Integer)
- created_at (Timestamp)
```

**Posts Table (Cassandra):**
```sql
- post_id (UUID, Primary Key)
- user_id (UUID)
- content (Text)
- media_urls (List)
- like_count (Integer)
- retweet_count (Integer)
- created_at (Timestamp)
```

**Follows Table:**
```sql
- follower_id (UUID)
- following_id (UUID)
- created_at (Timestamp)
- Primary Key (follower_id, following_id)
```

## **Phase 3: Feed Generation Strategy**

### **3.1 Push vs Pull Model**

**Push Model (Fan-out on Write):**
- Pre-compute feeds when posts are created
- Store in Redis cache per user
- Good for users with few followers

**Pull Model (Fan-out on Read):**
- Compute feed when user requests it
- Good for celebrities with millions of followers

**Hybrid Approach:**
- Push for normal users (< 1M followers)
- Pull for celebrities (> 1M followers)
- Cache frequently accessed feeds

### **3.2 Timeline Generation Algorithm**

```
1. Get user's following list
2. Fetch recent posts from followed users
3. Merge and rank by timestamp/relevance
4. Apply filters (blocked users, content type)
5. Paginate results
6. Cache for future requests
```

## **Phase 4: System Components**

### **4.1 Microservices Architecture**

**User Service:**
- User registration/authentication
- Profile management
- Follow/unfollow operations

**Post Service:**
- Create/delete posts
- Media upload handling
- Post metadata management

**Feed Service:**
- Timeline generation
- Feed ranking algorithms
- Cache management

**Notification Service:**
- Push notifications
- Email notifications
- Real-time updates

### **4.2 Data Storage Strategy**

**SQL Database (PostgreSQL):**
- User profiles
- Relationships (follows)
- Account settings

**NoSQL Database (Cassandra):**
- Posts and tweets
- User timelines
- Activity feeds

**Cache Layer (Redis):**
- User sessions
- Hot feeds
- Trending topics

**Object Storage (S3):**
- Media files (images, videos)
- User avatars
- Static assets

## **Phase 5: Scaling Considerations**

### **5.1 Read Scaling**
- Read replicas for databases
- CDN for media content
- Redis clusters for caching
- Horizontal service scaling

### **5.2 Write Scaling**
- Database sharding by user_id
- Message queues for async processing
- Write-through caching
- Event-driven architecture

### **5.3 Geographic Distribution**
- Multi-region deployment
- Data replication across regions
- Edge caching with CDN
- Regional load balancing

## **Phase 6: Implementation Roadmap**

### **Week 1-2: Foundation**
- Set up basic microservices
- Implement user authentication
- Create database schemas

### **Week 3-4: Core Features**
- Post creation/deletion
- Basic follow functionality
- Simple timeline generation

### **Week 5-6: Feed Optimization**
- Implement caching layer
- Add feed ranking algorithms
- Optimize database queries

### **Week 7-8: Advanced Features**
- Add media upload
- Implement notifications
- Add real-time updates

### **Week 9-10: Scaling & Performance**
- Load testing
- Performance optimization
- Add monitoring and metrics

## **Phase 7: Technology Stack**

**Backend:** Go/Java/Python
**Databases:** PostgreSQL, Cassandra, Redis
**Message Queue:** Apache Kafka/RabbitMQ
**API Gateway:** Kong/AWS API Gateway
**Monitoring:** Prometheus, Grafana
**Container:** Docker, Kubernetes
**Cloud:** AWS/GCP/Azure

This plan provides a comprehensive approach to building a scalable social media feed system similar to Twitter, focusing on both functional requirements and system design principles.

## **Phase 8: Go Backend Architecture & Libraries**

### **8.1 Recommended Go Libraries**

**Web Framework:**
```go
// Gin - High performance HTTP web framework
github.com/gin-gonic/gin

```

**Database Libraries:**
```go
// PostgreSQL
github.com/lib/pq              // PostgreSQL driver
github.com/jmoiron/sqlx        // Extensions to database/sql
gorm.io/gorm                   // ORM library
gorm.io/driver/postgres        // GORM PostgreSQL driver

// Cassandra
github.com/gocql/gocql         // Cassandra driver

// Redis
github.com/go-redis/redis/v8   // Redis client
```

**Message Queue:**
```go
// Kafka
github.com/segmentio/kafka-go  // Kafka library
github.com/confluentinc/confluent-kafka-go // Confluent Kafka

// RabbitMQ
github.com/streadway/amqp      // AMQP 0.9.1 client
```

**Authentication & Security:**
```go
github.com/golang-jwt/jwt/v4   // JWT tokens
golang.org/x/crypto/bcrypt     // Password hashing
github.com/google/uuid         // UUID generation
```

**Configuration & Environment:**
```go
github.com/spf13/viper         // Configuration management
github.com/joho/godotenv       // Load .env files
```

**Logging & Monitoring:**
```go
github.com/sirupsen/logrus     // Structured logging
go.uber.org/zap               // Fast, structured logging
github.com/prometheus/client_golang // Prometheus metrics
```

**Testing:**
```go
github.com/stretchr/testify    // Testing toolkit
github.com/golang/mock         // Mock generation
```

### **8.2 Project Structure (Go)**

```
twitter-service/
├── cmd/
│   └── server/
│       └── main.go                 // Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── post_handler.go
│   │   │   └── feed_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   └── logging.go
│   │   └── routes/
│   │       └── routes.go
│   ├── config/
│   │   └── config.go               // Configuration management
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── user.go
│   │   │   ├── post.go
│   │   │   └── follow.go
│   │   └── repositories/
│   │       ├── user_repository.go
│   │       ├── post_repository.go
│   │       └── feed_repository.go
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── postgres.go
│   │   │   ├── cassandra.go
│   │   │   └── redis.go
│   │   ├── messaging/
│   │   │   └── kafka.go
│   │   └── storage/
│   │       └── s3.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── post_service.go
│   │   └── feed_service.go
│   └── utils/
│       ├── jwt.go
│       ├── validation.go
│       └── response.go
├── pkg/
│   ├── logger/
│   │   └── logger.go
│   └── errors/
│       └── errors.go
├── migrations/
│   ├── postgres/
│   │   ├── 001_create_users.up.sql
│   │   └── 001_create_users.down.sql
│   └── cassandra/
│       └── keyspace.cql
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

### **8.3 Architecture Patterns**

**Clean Architecture:**
- **Domain Layer**: Entities and business logic
- **Application Layer**: Use cases and services
- **Infrastructure Layer**: Database, external APIs
- **Interface Layer**: HTTP handlers, middleware

**Repository Pattern:**
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
}
```

**Dependency Injection:**
```go
type Services struct {
    UserService UserService
    PostService PostService
    FeedService FeedService
}

type Dependencies struct {
    DB       *sql.DB
    Redis    *redis.Client
    Kafka    *kafka.Writer
    Config   *config.Config
}
```

### **8.4 Performance Optimizations**

**Connection Pooling:**
```go
// PostgreSQL pool configuration
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Redis pool
rdb := redis.NewClient(&redis.Options{
    PoolSize:     10,
    MinIdleConns: 5,
})
```

**Caching Strategy:**
```go
// Cache-aside pattern
func (s *PostService) GetPost(ctx context.Context, id string) (*Post, error) {
    // Try cache first
    cached, err := s.cache.Get(ctx, "post:"+id)
    if err == nil {
        return cached, nil
    }
    
    // Fetch from database
    post, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.Set(ctx, "post:"+id, post, time.Hour)
    return post, nil
}
```

**Goroutines for Async Processing:**
```go
// Fan-out feed generation
func (s *FeedService) FanOutPost(ctx context.Context, post *Post) {
    followers, _ := s.userRepo.GetFollowers(ctx, post.UserID)
    
    semaphore := make(chan struct{}, 100) // Limit concurrency
    var wg sync.WaitGroup
    
    for _, followerID := range followers {
        wg.Add(1)
        go func(fID string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            s.addToUserFeed(ctx, fID, post)
        }(followerID)
    }
    wg.Wait()
}
```

### **8.5 Monitoring & Observability**

**Prometheus Metrics:**
```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
            Help: "Database query duration",
        },
        []string{"query_type"},
    )
)
```

**Structured Logging:**
```go
logger := logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "request_id": requestID,
    "operation": "create_post",
})
logger.Info("Post created successfully")
```

### **8.6 Docker Configuration**

**Multi-stage Dockerfile:**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

This Go-specific architecture provides a solid foundation for building a scalable, maintainable Twitter-like social media system with proper separation of concerns and Go best practices.

## **📋 Implementation Checklist - Step by Step**

### **Phase 1: Database Setup (Users Table Only)**
- [x] Create simple Docker Compose for PostgreSQL
- [x] Create users table migration
- [x] Set up basic environment configuration

### **Phase 2: Go Application Setup (Current)**
- [x] Initialize Go modules
- [x] Create basic Gin server with routes
- [x] Add GORM connection to PostgreSQL
- [x] Create User entity/model
- [x] Create Google OAuth integration
- [x] Add JWT authentication
- [x] Create auth handlers and middleware
- [x] Create HTML test page
- [ ] Test Google OAuth flow end-to-end

### **Phase 3: User API Endpoints (Future)**
- [ ] Add user profile update endpoints
- [ ] Add user search functionality
- [ ] Add user follow/unfollow endpoints

### **Week 3-4: User Management**
- [ ] Create User entity/model
- [ ] Implement user registration endpoint
- [ ] Implement user login endpoint
- [ ] Add JWT authentication middleware
- [ ] Create user profile endpoints (GET, PUT)
- [ ] Add password hashing (bcrypt)
- [ ] Implement user validation
- [ ] Add unit tests for user service

### **Week 5-6: Post System**
- [ ] Create Post entity/model
- [ ] Implement create post endpoint
- [ ] Implement get posts endpoint
- [ ] Implement delete post endpoint
- [ ] Add post validation (character limits)
- [ ] Add post media upload (basic)
- [ ] Create post repository layer
- [ ] Add unit tests for post service

### **Week 7-8: Follow System**
- [ ] Create Follow entity/model
- [ ] Implement follow user endpoint
- [ ] Implement unfollow user endpoint
- [ ] Implement get followers endpoint
- [ ] Implement get following endpoint
- [ ] Add follow validation (prevent self-follow)
- [ ] Update follower/following counts
- [ ] Add unit tests for follow service

### **Week 9-10: Basic Feed**
- [ ] Create basic timeline generation
- [ ] Implement get user timeline endpoint
- [ ] Implement get home feed endpoint (pull model)
- [ ] Add pagination for feeds
- [ ] Add basic feed sorting (by timestamp)
- [ ] Create feed repository layer
- [ ] Add Redis for basic caching
- [ ] Add unit tests for feed service

### **Week 11-12: Like & Interaction System**
- [ ] Create Like entity/model
- [ ] Implement like post endpoint
- [ ] Implement unlike post endpoint
- [ ] Add like count to posts
- [ ] Implement get user likes endpoint
- [ ] Add like validation (prevent duplicate likes)
- [ ] Update post like counts
- [ ] Add unit tests for like service

### **Week 13-14: Testing & Documentation**
- [ ] Add integration tests
- [ ] Add API documentation (Swagger)
- [ ] Add database migrations
- [ ] Add logging middleware
- [ ] Add error handling middleware
- [ ] Add rate limiting middleware
- [ ] Create deployment scripts
- [ ] Performance testing with basic load

### **Future Enhancements (Phase 2)**
- [ ] Add Cassandra for posts scaling
- [ ] Implement push model for feed generation
- [ ] Add real-time notifications
- [ ] Add media upload to S3
- [ ] Add search functionality
- [ ] Add trending topics
- [ ] Add comment system
- [ ] Add retweet functionality

---

## **🛠️ Technology Stack (Confirmed)**
- **Framework**: Gin (github.com/gin-gonic/gin)
- **ORM**: GORM (gorm.io/gorm)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Auth**: JWT (github.com/golang-jwt/jwt/v4)
- **Password**: bcrypt (golang.org/x/crypto/bcrypt)
- **Config**: Viper (github.com/spf13/viper)
- **Logging**: Logrus (github.com/sirupsen/logrus)
- **Testing**: Testify (github.com/stretchr/testify)