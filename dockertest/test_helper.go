package dockertest

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"
)

//atexit := atexit.NewOnExit()
//atexit.Add(func() {
//	dockertest.KillAll()
//})
//atexit.Exit(testMain(m))

//func WrapCleanup

var resources = []*dockertest.Resource{}
var pool *dockertest.Pool

func KillAllTestDatabases() {
	for _, r := range resources {
		pool.Purge(r)
	}
}

func Register() *OnExit {
	onexit := NewOnExit()
	onexit.Add(func() {
		KillAllTestDatabases()
	})
	return onexit
}

func ConnectToTestPostgreSQL() (*sqlx.DB, error) {
	if url := os.Getenv("TEST_DATABASE_POSTGRESQL"); url != "" {
		log.Println("Found postgresql test database config, skipping dockertest...")
		db, err := sqlx.Open("postgres", url)
		if err != nil {
			return nil, errors.Wrap(err, "Could not connect to bootstrapped database")
		}
		return db, nil
	}

	var db *sqlx.DB
	var err error

	pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, errors.Wrap(err, "Could not connect to docker")
	}

	resource, err := pool.Run("postgres", "9.6", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=hydra"})
	if err != nil {
		return nil, errors.Wrap(err, "Could not start resource")
	}

	if err = pool.Retry(func() error {
		var err error
		db, err = sqlx.Open("postgres", fmt.Sprintf("postgres://postgres:secret@localhost:%s/hydra?sslmode=disable", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		pool.Purge(resource)
		return nil, errors.Wrap(err, "Could not connect to docker")
	}
	resources = append(resources, resource)

	return db, nil
}

func ConnectToTestMySQL() (*sqlx.DB, error) {
	if url := os.Getenv("TEST_DATABASE_MYSQL"); url != "" {
		log.Println("Found mysql test database config, skipping dockertest...")
		db, err := sqlx.Open("mysql", url)
		if err != nil {
			return nil, errors.Wrap(err, "Could not connect to bootstrapped database")
		}
		return db, nil
	}

	var db *sqlx.DB
	var err error

	pool, err = dockertest.NewPool("")
	pool.MaxWait = time.Minute * 5
	if err != nil {
		return nil, errors.Wrap(err, "Could not connect to docker")
	}

	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		return nil, errors.Wrap(err, "Could not start resource")
	}

	if err = pool.Retry(func() error {
		var err error
		db, err = sqlx.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql?parseTime=true", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		pool.Purge(resource)
		return nil, errors.Wrap(err, "Could not connect to docker")
	}
	resources = append(resources, resource)

	return db, nil
}
