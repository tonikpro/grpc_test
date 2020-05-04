package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"testing"

	driverSql "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tonikpro/grpc_test/repository"
	"github.com/tonikpro/grpc_test/service"
	"github.com/tonikpro/grpc_test/test"
	"github.com/tonikpro/grpc_test/test/models"
	mygrpc "github.com/tonikpro/grpc_test/transport/grpc"
	"github.com/tonikpro/grpc_test/transport/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var (
	db *sqlx.DB
)

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

func TestGrpcSuite(t *testing.T) {
	mysuite := &testSuite{
		DB: db,
	}
	suite.Run(t, mysuite)
}

type testSuite struct {
	suite.Suite
	DB  *sqlx.DB
	m   *test.Migration
	lis *bufconn.Listener
}

func (s *testSuite) SetupSuite() {
	log.Println("main_test: Starting a Test. Migrating the Database")
	DisableLogging()
	var err error
	s.m, err = test.NewMigration(s.DB, "test/migrations")
	require.NoError(s.T(), err)
	_, err = s.m.Up()
	require.NoError(s.T(), err)
	log.Println("main_test: Database Migrated Successfully")
	log.Println("main_test: Creating Test data in database")
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
	log.Println("main_test: Finish creating test data in database")

	repo := repository.NewA2billingRepository(s.DB)
	svc := service.NewService(repo)
	ep := service.NewA2billingEndpoints(svc)
	s.lis = bufconn.Listen(bufSize)
	sr := grpc.NewServer()
	pb.RegisterA2BillingServer(sr, mygrpc.NewGrpcServer(ep))
	go func() {
		if err := sr.Serve(s.lis); err != nil {
			log.Fatalf("main_test: Server exited with error: %v", err)
		}
	}()
}

func (s *testSuite) TearDownTest() {
	log.Println("main_test: Finishing Test. Dropping The Database")
	_, err := s.m.Down()
	require.NoError(s.T(), err)
	log.Println("main_test: Database Dropped Successfully")
}

func (s *testSuite) bufDialer(context.Context, string) (net.Conn, error) {
	return s.lis.Dial()
}

func (s *testSuite) TestGrpcCall_GetAgentIdsByParentAgentID() {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(s.bufDialer), grpc.WithInsecure())
	if err != nil {
		s.T().Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewA2BillingClient(conn)
	type args struct {
		ctx context.Context
		id  *pb.GetAgentIdsByParentAgentIDRequest
	}
	tests := []struct {
		name       string
		s          pb.A2BillingClient
		args       args
		wantResult *pb.GetAgentIdsByParentAgentIDResponse
		wantErr    bool
	}{
		{
			"agent_id=1",
			client,
			args{
				ctx,
				&pb.GetAgentIdsByParentAgentIDRequest{AgentId: 1},
			},
			&pb.GetAgentIdsByParentAgentIDResponse{AgentIds: []int32{5, 6}},
			false,
		},
		{
			"agent_id=2",
			client,
			args{
				ctx,
				&pb.GetAgentIdsByParentAgentIDRequest{AgentId: 2},
			},
			&pb.GetAgentIdsByParentAgentIDResponse{AgentIds: []int32{7, 8}},
			false,
		},
		{
			"agent_id=3",
			client,
			args{
				ctx,
				&pb.GetAgentIdsByParentAgentIDRequest{AgentId: 3},
			},
			&pb.GetAgentIdsByParentAgentIDResponse{AgentIds: []int32{9}},
			false,
		},
		{
			"agent_id=4",
			client,
			args{
				ctx,
				&pb.GetAgentIdsByParentAgentIDRequest{AgentId: 4},
			},
			&pb.GetAgentIdsByParentAgentIDResponse{AgentIds: []int32{10}},
			false,
		},
		{
			"agent_id=0",
			client,
			args{
				ctx,
				&pb.GetAgentIdsByParentAgentIDRequest{AgentId: 0},
			},
			&pb.GetAgentIdsByParentAgentIDResponse{AgentIds: []int32{1, 2, 3, 4}},
			false,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.s.GetAgentIdsByParentAgentID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("grpc.GetAgentIdsByParentAgentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult.AgentIds, tt.wantResult.AgentIds) {
				t.Errorf("grpc.GetAgentIdsByParentAgentID() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

type NopLogger struct {
}

func (l NopLogger) Print(v ...interface{}) {}

func DisableLogging() {
	nopLogger := NopLogger{}
	driverSql.SetLogger(nopLogger)
}
