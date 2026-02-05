# 📌 Pastebin System Design Plan

## 1. Requirements Gathering

### Functional Requirements

- [ ] User can paste text (up to N MB)
- [ ] System generates a unique URL (short hash)
- [ ] Anyone with the link can retrieve the paste

#### Optional Features:
- [ ] Paste expiration (e.g., after 24 hours)
- [ ] Private vs public pastes
- [ ] Syntax highlighting

### Non-Functional Requirements

> **Performance & Reliability Goals**

- **High availability**: service should not lose data
- **Low latency**: retrieving a paste should be fast (<100ms ideally)
- **Scalability**: support millions of pastes/day
- **Storage efficiency**: don't store duplicates unnecessarily

## 2. API Design

### Endpoints

#### Create Paste
- **POST** `/paste` → create a new paste
  - **Input**: `{ content, expiration_time }`
  - **Output**: `{ paste_id, url }`

#### Retrieve Paste
- **GET** `/paste/{paste_id}` → retrieve paste
  - **Output**: `{ content, created_at, expires_at }`

## 3. High-Level Architecture

### Core Components

1. **API Gateway / Load Balancer** → entry point for requests
2. **Application Servers** → handle business logic
3. **Database** → store pastes
   - **Key**: `paste_id`
   - **Value**: `content, created_at, expires_at`
4. **Cache** (Redis / Memcached) → serve frequently accessed pastes
5. **Background Workers** → cleanup expired pastes

## 4. Data Modeling

### Database Choice

| Scale | Database Type | Examples |
|-------|---------------|----------|
| Small scale | Relational DB | PostgreSQL, MySQL |
| Large scale | NoSQL | Cassandra, DynamoDB, MongoDB |

### Schema (SQL style)

```sql
CREATE TABLE pastes (
    paste_id VARCHAR(8) PRIMARY KEY,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NULL
);
```

## 5. Paste ID Generation

### Options

| Option | Method | Pros | Cons |
|--------|--------|------|------|
| **Option 1** | Base62 encoding of auto-increment ID | Simpler implementation | Predictable IDs |
| **Option 2** | Random hash (MD5/SHA1 truncated) | Harder to guess, secure | More complex |

**Example**: `0001` → `"aZ"`

### Trade-offs
- **Sequential IDs** → simpler, but predictable
- **Random IDs** → harder to guess, good for security

## 6. Expiration & Cleanup

### Strategy
- Store expiration timestamp in DB

### Approaches

#### 1. Lazy Deletion
- ✅ Check expiry at read time
- ✅ Return 404 if expired
- ⚡ Simple implementation

#### 2. Active Deletion  
- ✅ Background job scans expired pastes
- ✅ Removes them proactively
- ⚡ Better resource management

## 7. Scaling Considerations

### Read-Heavy Workload
> **Reads > Writes** - Add caching layer (Redis)

- Popular pastes stay in memory

### Sharding Strategy
> For billions of pastes

- **By paste_id hash** → Ensures distribution across DB nodes

### Content Size Management
- For very large pastes:
  - Store text in **object storage** (S3, MinIO)
  - Keep metadata in DB

## 8. Security

### Protection Measures
- [ ] **Limit paste size** to prevent abuse (e.g., max 10 MB)
- [ ] **Rate limiting** (avoid spam/bots)
- [ ] **Optional encryption** for private pastes

## 9. Monitoring & Logging

### Key Metrics
- 📊 Request counts
- ⏱️ Latency
- 💾 Cache hit/miss ratio

### Logging
- ❌ Error logs for failed DB/cache operations

---

# 📈 Step-by-Step Challenge Plan

## Week 1 – Core MVP
- [ ] Implement **POST** `/paste` and **GET** `/paste/{id}`
- [ ] Use in-memory map (Go, Python, or Node)
- [ ] Generate random IDs

## Week 2 – Add Persistence
- [ ] Replace in-memory map with **PostgreSQL** or **SQLite**
- [ ] Add expiration support

## Week 3 – Caching & Scaling
- [ ] Integrate **Redis** for caching
- [ ] Implement lazy vs active expiration cleanup

## Week 4 – Advanced Features
- [ ] Add paste expiration policies:
  - ⏰ 10 minutes
  - 📅 1 day  
  - 📆 1 week
- [ ] Add syntax highlighting
- [ ] Add private (password-protected) pastes


