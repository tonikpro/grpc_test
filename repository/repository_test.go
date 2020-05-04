package repository

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	driverSql "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tonikpro/grpc_test/test"
	"github.com/tonikpro/grpc_test/test/models"
)

var db *sqlx.DB

type mysqlSuite struct {
	suite.Suite
	DB        *sqlx.DB
	Migration *test.Migration
}

func (s *mysqlSuite) SetupSuite() {
	log.Println("Starting a Test. Migrating the Database")
	DisableLogging()
	var err error
	s.Migration, err = test.NewMigration(s.DB, "../test/migrations")
	require.NoError(s.T(), err)
	_, err = s.Migration.Up()
	require.NoError(s.T(), err)
	log.Println("Database Migrated Successfully")
	log.Println("Creating Test data in database")
	agents := []models.Agent{
		{
			AgentID:       1,
			ParentAgentID: 0,
		},
		{
			AgentID:       2,
			ParentAgentID: 0,
		},
		{
			AgentID:       3,
			ParentAgentID: 0,
		},
		{
			AgentID:       4,
			ParentAgentID: 0,
		},
		{
			AgentID:       5,
			ParentAgentID: 1,
		},
		{
			AgentID:       6,
			ParentAgentID: 1,
		},
		{
			AgentID:       7,
			ParentAgentID: 2,
		},
		{
			AgentID:       8,
			ParentAgentID: 2,
		},
		{
			AgentID:       9,
			ParentAgentID: 3,
		},
		{
			AgentID:       10,
			ParentAgentID: 4,
		},
	}
	query := `INSERT INTO cc_agent (id, parent_agent_id) VALUES (?,?)`
	stmt, err := s.DB.Prepare(query)
	require.NoError(s.T(), err)
	defer Close(stmt)
	for _, agent := range agents {
		_, err := stmt.Exec(agent.AgentID, agent.ParentAgentID)
		require.NoError(s.T(), err)
	}
	log.Println("Finish creating test data in database")
}

func (s *mysqlSuite) TearDownTest() {
	log.Println("Finishing Test. Dropping The Database")
	_, err := s.Migration.Down()
	require.NoError(s.T(), err)
	log.Println("Database Dropped Successfully")
}

func (s *mysqlSuite) TestGetAgentIdsByParentAgentID() {
	type args struct {
		id int32
	}

	repo := NewA2billingRepository(s.DB)
	tests := []struct {
		name       string
		r          A2billingRepository
		args       args
		wantResult []int32
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			"parent 1",
			repo,
			args{
				1,
			},
			[]int32{5, 6},
			false,
		},
		{
			"parent 2",
			repo,
			args{
				2,
			},
			[]int32{7, 8},
			false,
		},
		{
			"parent 3",
			repo,
			args{
				3,
			},
			[]int32{9},
			false,
		},
		{
			"parent 4",
			repo,
			args{
				4,
			},
			[]int32{10},
			false,
		},
		{
			"parent 5",
			repo,
			args{
				5,
			},
			[]int32{},
			false,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.r.GetAgentIdsByParentAgentID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("a2billingRepository.GetAgentIdsByParentAgentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotResult) != 0 && len(tt.wantResult) != 0 {
				if !reflect.DeepEqual(gotResult, tt.wantResult) {
					t.Errorf("a2billingRepository.GetAgentIdsByParentAgentID() = %v, want %v", gotResult, tt.wantResult)
				}
			}
		})
	}
}

func DisableLogging() {
	nopLogger := NopLogger{}
	driverSql.SetLogger(nopLogger)
}

type NopLogger struct {
}

func (l NopLogger) Print(v ...interface{}) {}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func TestRepositorySuite(t *testing.T) {
	mysuite := &mysqlSuite{
		DB: db,
	}
	suite.Run(t, mysuite)
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sqlx.Connect("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
