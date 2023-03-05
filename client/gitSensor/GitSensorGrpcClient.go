package gitSensor

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	pb "github.com/devtron-labs/protos/git-sensor"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

const (
	ContextTimeoutInSeconds = 10
	MaxMsgSizeBytes         = 20 * 1024 * 1024
)

type GitSensorGrpcClient interface {
	GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (*pb.GitChanges, error)
}

type GitSensorGrpcClientImpl struct {
	logger        *zap.SugaredLogger
	config        *GitSensorGrpcClientConfig
	serviceClient pb.GitSensorServiceClient
}

func NewGitSensorGrpcClientImpl(logger *zap.SugaredLogger, config *GitSensorGrpcClientConfig) *GitSensorGrpcClientImpl {
	return &GitSensorGrpcClientImpl{
		logger: logger,
		config: config,
	}
}

// getGitSensorServiceClient initializes and returns gRPC GitSensorService client
func (client *GitSensorGrpcClientImpl) getGitSensorServiceClient() (pb.GitSensorServiceClient, error) {
	if client.serviceClient == nil {
		conn, err := client.getConnection()
		if err != nil {
			return nil, err
		}
		client.serviceClient = pb.NewGitSensorServiceClient(conn)
	}
	return client.serviceClient, nil
}

// getConnection initializes and returns a grpc client connection
func (client *GitSensorGrpcClientImpl) getConnection() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeoutInSeconds*time.Second)
	defer cancel()

	// Configure gRPC dial options
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxMsgSizeBytes),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	endpoint := fmt.Sprintf("dns:///%s", client.config.Url)

	// initialize connection
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		client.logger.Errorw("error while initializing grpc connection",
			"endpoint", endpoint,
			"err", err)
		return nil, err
	}
	return conn, nil
}

type GitSensorGrpcClientConfig struct {
	Url string `env:"GIT_SENSOR_HOST" envDefault:"127.0.0.1:7070"`
}

// GetConfig parses and returns GitSensor gRPC client configuration
func GetConfig() (*GitSensorGrpcClientConfig, error) {
	cfg := &GitSensorGrpcClientConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (client *GitSensorGrpcClientImpl) GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (
	*pb.GitChanges, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetChangesInRelease(ctx, req)
}
