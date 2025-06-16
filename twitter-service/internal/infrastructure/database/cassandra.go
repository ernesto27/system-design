package database

import (
	"fmt"
	"log"
	"twitterservice/internal/config"

	"github.com/gocql/gocql"
)

type CassandraDB struct {
	Session *gocql.Session
}

func NewCassandraConnection(cfg *config.Config) (*CassandraDB, error) {
	// First, connect without keyspace to create it if needed
	cluster := gocql.NewCluster(cfg.Cassandra.Host)
	cluster.Port = cfg.Cassandra.Port
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = cfg.Cassandra.Timeout

	// Create session without keyspace
	tempSession, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra: %w", err)
	}

	// Create keyspace if it doesn't exist
	createKeyspaceQuery := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH REPLICATION = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}`, cfg.Cassandra.Keyspace)

	if err := tempSession.Query(createKeyspaceQuery).Exec(); err != nil {
		tempSession.Close()
		return nil, fmt.Errorf("failed to create keyspace: %w", err)
	}

	log.Printf("Keyspace '%s' created or already exists", cfg.Cassandra.Keyspace)

	// Close temporary session
	tempSession.Close()

	// Now connect with the keyspace
	cluster.Keyspace = cfg.Cassandra.Keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra with keyspace: %w", err)
	}

	// Create tables
	if err := createTables(session); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Successfully connected to Cassandra and initialized schema")

	return &CassandraDB{
		Session: session,
	}, nil
}

func createTables(session *gocql.Session) error {
	// Create posts table (without DEFAULT values - not supported in Cassandra)
	createPostsTable := `
		CREATE TABLE IF NOT EXISTS posts (
			id UUID PRIMARY KEY,
			user_id UUID,
			content TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			is_deleted BOOLEAN
		)`

	if err := session.Query(createPostsTable).Exec(); err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	// Create indexes
	createUserIDIndex := `CREATE INDEX IF NOT EXISTS posts_user_id_idx ON posts (user_id)`
	if err := session.Query(createUserIDIndex).Exec(); err != nil {
		return fmt.Errorf("failed to create user_id index: %w", err)
	}

	createTimeIndex := `CREATE INDEX IF NOT EXISTS posts_created_at_idx ON posts (created_at)`
	if err := session.Query(createTimeIndex).Exec(); err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}

	// Create user timeline table
	createTimelineTable := `
		CREATE TABLE IF NOT EXISTS user_timeline (
			user_id UUID,
			post_id UUID,
			created_at TIMESTAMP,
			content TEXT,
			author_id UUID,
			PRIMARY KEY (user_id, created_at, post_id)
		) WITH CLUSTERING ORDER BY (created_at DESC, post_id ASC)`

	if err := session.Query(createTimelineTable).Exec(); err != nil {
		return fmt.Errorf("failed to create user_timeline table: %w", err)
	}

	log.Println("Cassandra tables and indexes created successfully")
	return nil
}

func (db *CassandraDB) Close() {
	if db.Session != nil {
		db.Session.Close()
	}
}
