package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"github.com/tonikpro/grpc_test/repository"
	"github.com/tonikpro/grpc_test/service"
	mygrpc "github.com/tonikpro/grpc_test/transport/grpc"
	"github.com/tonikpro/grpc_test/transport/grpc/pb"
	"google.golang.org/grpc"
)

func main() {
	var (
		mysqlConnectionString = flag.String("mysql.conn", "", "MySQL connection string")
		tcpPort               = flag.Int("tcp.port", 50000, "TCP port for gRpc connections")
	)

	flag.Parse()
	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	errc := make(chan error)

	logger.Log("msg", "hello")
	defer logger.Log("msg", "goodbye")
	db, err := sqlx.Connect("mysql", *mysqlConnectionString)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	svc := service.NewService(repository.NewA2billingRepository(db))

	endpoints := service.NewA2billingEndpoints(svc)
	endpoints.GetAgentIdsByParentAgentID = service.LoggingMiddleware(logger)(endpoints.GetAgentIdsByParentAgentID)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *tcpPort))
		if err != nil {
			errc <- err
			return
		}
		s := grpc.NewServer()
		pb.RegisterA2BillingServer(s, mygrpc.NewGrpcServer(endpoints))
		if err := s.Serve(lis); err != nil {
			errc <- err
			return
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()
	logger.Log("error", <-errc)
}
