package orm_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iMerica/jango/orm"
)

type TestAuthor struct {
	ID    uint   `jango:"primary_key"`
	Name  string `jango:"type:char,max_length:100"`
	Email string `jango:"unique"`
}

type TestTag struct {
	ID   uint   `jango:"primary_key"`
	Name string `jango:"unique"`
}

type TestArticle struct {
	ID        uint       `jango:"primary_key"`
	Title     string     `jango:"type:char,max_length:200"`
	Body      string     `jango:"type:text"`
	Published bool       `jango:"type:boolean,default:false"`
	AuthorID  uint       `jango:"column:author_id"` // Simplified foreign key for MVP test
}

func init() {
	orm.RegisterModel("integration", &TestAuthor{})
	orm.RegisterModel("integration", &TestTag{})
	orm.RegisterModel("integration", &TestArticle{})
}

func setupTestDB(t *testing.T) *orm.DB {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		os.Setenv("DATABASE_URL", "postgres://jango:password@localhost:5432/jango_test?sslmode=disable")
	}

	config := orm.DefaultDBConfig()

	ctx := context.Background()
	
	// Fast connection timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := orm.OpenDB(ctxTimeout, config)
	if err != nil {
		t.Skipf("Skipping integration test; unable to connect to DB: %v", err)
	}

	// Create tables
	queries := []string{
		`DROP TABLE IF EXISTS integration_testarticle CASCADE;`,
		`DROP TABLE IF EXISTS integration_testtag CASCADE;`,
		`DROP TABLE IF EXISTS integration_testauthor CASCADE;`,
		
		`CREATE TABLE integration_testauthor (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email TEXT UNIQUE NOT NULL
		);`,
		
		`CREATE TABLE integration_testtag (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);`,
		
		`CREATE TABLE integration_testarticle (
			id SERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			body TEXT NOT NULL,
			published BOOLEAN DEFAULT false,
			author_id BIGINT REFERENCES integration_testauthor(id) ON DELETE CASCADE
		);`,
	}

	for _, q := range queries {
		_, err := conn.Exec(ctx, q)
		if err != nil {
			t.Fatalf("Failed to execute setup query: %s\nErr: %v", q, err)
		}
	}

	orm.SetDefaultDB(conn)
	return conn
}

func TestIntegrationCRUDAndFilters(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close()

	ctx := context.Background()

	t.Run("Create and Get", func(t *testing.T) {
		author := &TestAuthor{Name: "John Doe", Email: "john@example.com"}
		err := orm.Objects[TestAuthor]("integration", "TestAuthor").Create(ctx, author)
		if err != nil {
			t.Fatalf("Failed to create author: %v", err)
		}

		fetched, err := orm.Objects[TestAuthor]("integration", "TestAuthor").Get(ctx, orm.L("email", "john@example.com"))
		if err != nil {
			t.Fatalf("Failed to get author: %v", err)
		}
		if fetched.Name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got %s", fetched.Name)
		}
	})

	t.Run("Filter and Exclude", func(t *testing.T) {
		orm.Objects[TestTag]("integration", "TestTag").Create(ctx, &TestTag{Name: "Go"})
		orm.Objects[TestTag]("integration", "TestTag").Create(ctx, &TestTag{Name: "Python"})
		orm.Objects[TestTag]("integration", "TestTag").Create(ctx, &TestTag{Name: "Rust"})

		tags, err := orm.Objects[TestTag]("integration", "TestTag").Filter(orm.L("name__contains", "o")).AllRecords(ctx)
		if err != nil {
			t.Fatalf("Filter failed: %v", err)
		}
		if len(tags) != 2 { // Go, Python
			t.Errorf("Expected 2 tags containing 'o', got %d", len(tags))
		}

		tags, err = orm.Objects[TestTag]("integration", "TestTag").Exclude(orm.L("name", "Go")).OrderBy("name").AllRecords(ctx)
		if err != nil {
			t.Fatalf("Exclude failed: %v", err)
		}
		if len(tags) != 2 || tags[0].Name != "Python" {
			t.Errorf("Exclude/OrderBy unexpected result: %v", tags)
		}
	})

	t.Run("Aggregate", func(t *testing.T) {
		count, err := orm.Objects[TestTag]("integration", "TestTag").Count(ctx)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count < 3 {
			t.Errorf("Expected at least 3 tags, got %d", count)
		}
	})

	t.Run("Unique Constraint", func(t *testing.T) {
		err := orm.Objects[TestAuthor]("integration", "TestAuthor").Create(ctx, &TestAuthor{Name: "Duplicate", Email: "john@example.com"})
		if err == nil {
			t.Error("Expected error when violating unique constraint")
		}
	})
	
	t.Run("Q Objects OR Logic", func(t *testing.T) {
	    // Go or Rust
	    qNode := orm.QOr(orm.Q(orm.L("name", "Go")), orm.Q(orm.L("name", "Rust")))
	    
	    tags, err := orm.Objects[TestTag]("integration", "TestTag").FilterQ(qNode).AllRecords(ctx)
		if err != nil {
			t.Fatalf("FilterQ failed: %v", err)
		}
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags for QOr logic, got %d", len(tags))
		}
	})
}
