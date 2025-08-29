package server

import (
	"Base_node/pb"
	"context"
	"log"
)

type FederationServer struct {
	pb.UnimplementedBaseFederationServiceServer
	localRegion string
	baseNode    *BaseNodeServer
}

func NewFederationServer(region string, base *BaseNodeServer) *FederationServer {
	return &FederationServer{
		localRegion: region,
		baseNode:    base,
	}
}

func (s *FederationServer) RequestRemoteSuperNodes(ctx context.Context, req *pb.RemoteSuperRequest) (*pb.RemoteSuperResponse, error) {
	supers := s.baseNode.GetFilteredSuperNodes(req.Count, req.RequiredBandWidthMbps, req.MaxLatencyMs)

	var nodes []*pb.SuperNodeInfo
	for _, n := range supers {
		nodes = append(nodes, &pb.SuperNodeInfo{
			NodeId:             n.NodeID,
			Ip:                 n.IP,
			Port:               n.Port,
			Region:             n.Region,
			AvgLatencyMs:       n.AvgLatency,
			ExitPeersAvailable: 10,
			BandWidthMbps:      n.BandwidthMbps,
		})
	}

	log.Printf("🚀 Sending %d Super Nodes to supernode", len(nodes))

	return &pb.RemoteSuperResponse{
		SuperNodes: nodes,
	}, nil
}
