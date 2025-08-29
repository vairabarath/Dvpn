package client

import (
	"Base_node/pb"
	"context"
	"time"

	"google.golang.org/grpc"
)

func FetchRemoteSupers(addr string, region string, count int32, minBW float32, maxLatency float32) ([]*pb.SuperNodeInfo, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewBaseFederationServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.RequestRemoteSuperNodes(ctx, &pb.RemoteSuperRequest{
		TargetRegion:          region,
		Count:                 count,
		RequiredBandWidthMbps: minBW,
		MaxLatencyMs:          maxLatency,
	})
	if err != nil {
		return nil, err
	}

	return resp.SuperNodes, nil
}
