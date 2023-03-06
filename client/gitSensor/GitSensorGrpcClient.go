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
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"time"
)

const (
	ContextTimeoutInSeconds = 10
	MaxMsgSizeBytes         = 20 * 1024 * 1024
)

type GitSensorGrpcClient interface {
	GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (*GitChanges, error)
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
	Url string `env:"GIT_SENSOR_URL" envDefault:"127.0.0.1:7070"`
}

// GetConfig parses and returns GitSensor gRPC client configuration
func GetConfig() (*GitSensorGrpcClientConfig, error) {
	cfg := &GitSensorGrpcClientConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (client *GitSensorGrpcClientImpl) GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (
	*GitChanges, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}

	res, err := serviceClient.GetChangesInRelease(ctx, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// map res
	return client.mapGitChanges(res), nil
}

// mapGitChanges maps GitChanges from protobuf specified-type to golang struct type
func (client *GitSensorGrpcClientImpl) mapGitChanges(gitChanges *pb.GitChanges) *GitChanges {
	commits := make([]*Commit, 0)
	for _, item := range gitChanges.Commits {

		commit := &Commit{}

		// Map Hash
		if item.Hash != nil {
			commit.Hash = &Hash{
				Long:  item.Hash.Long,
				Short: item.Hash.Short,
			}
		}

		// Map Tree
		if item.Tree != nil {
			commit.Tree = &Tree{
				Long:  item.Tree.Long,
				Short: item.Tree.Short,
			}
		}

		// Map Author
		if item.Author != nil {
			commit.Author = &Author{
				Name:  item.Author.Name,
				Email: item.Author.Email,
			}
			if item.Author.Date != nil {
				commit.Author.Date = item.Author.Date.AsTime()
			}
		}

		// Map Committer
		if item.Committer != nil {
			commit.Committer = &Committer{
				Name:  item.Committer.Name,
				Email: item.Committer.Email,
			}
			if item.Committer.Date != nil {
				commit.Committer.Date = item.Committer.Date.AsTime()
			}
		}

		// Map Tag
		if item.Tag != nil {
			commit.Tag = &Tag{
				Name: item.Tag.Name,
			}
			if item.Tag.Date != nil {
				commit.Tag.Date = item.Tag.Date.AsTime()
			}
		}

		commit.Body = item.Body
		commit.Subject = item.Subject
		commits = append(commits, commit)
	}

	// Map FileStats
	fileStats := make([]object.FileStat, 0)
	for _, item := range gitChanges.FileStats {

		fileStats = append(fileStats, object.FileStat{
			Name:     item.Name,
			Addition: int(item.Addition),
			Deletion: int(item.Deletion),
		})
	}

	return &GitChanges{
		Commits:   commits,
		FileStats: fileStats,
	}
}
