package clients

import (
	pb "github.com/chgrape/storage-app/shared/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MediaClient struct {
	conn   *grpc.ClientConn
	Client pb.MediaClient
}

func NewMediaClient(addr string) (*MediaClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := pb.NewMediaClient(conn)

	return &MediaClient{
		conn:   conn,
		Client: c,
	}, nil
}

func (m *MediaClient) Close() {
	m.conn.Close()
}
