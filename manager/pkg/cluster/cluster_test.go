/*
Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	container "cloud.google.com/go/container/apiv1"
	"github.com/golang/protobuf/proto"
	containerpb "google.golang.org/genproto/googleapis/container/v1"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/api/option"

	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ = io.EOF
var _ = ptypes.MarshalAny
var _ status.Status

type mockClusterManagerServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	containerpb.ClusterManagerServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockClusterManagerServer) GetCluster(ctx context.Context, req *containerpb.GetClusterRequest) (*containerpb.Cluster, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*containerpb.Cluster), nil
}

func (s *mockClusterManagerServer) GetServerConfig(ctx context.Context, req *containerpb.GetServerConfigRequest) (*containerpb.ServerConfig, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*containerpb.ServerConfig), nil
}

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockClusterManager mockClusterManagerServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	containerpb.RegisterClusterManagerServer(serv, &mockClusterManager)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Print(err)
	}
	go serv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Print(err)
	}
	clientOpt = option.WithGRPCConn(conn)

	os.Exit(m.Run())
}

func TestLatestMasterVersionForReleaseSeries(t *testing.T) {
	var validMasterVersions = []string{
		"1.10.1-gke.5",
		"1.9.2-gke.1",
		"1.9.2-gke.0",
		"1.9.1-gke.0",
	}
	var expectedResponse = &containerpb.ServerConfig{
		ValidMasterVersions: validMasterVersions,
	}

	var expectedClusterResponse = &containerpb.Cluster{
		CurrentMasterVersion: "1.9.2-gke.1",
	}

	client, _ := container.NewClusterManagerClient(context.Background(), clientOpt)

	mockClusterManager.err = nil
	mockClusterManager.reqs = nil

	mockClusterManager.resps = append(mockClusterManager.resps[:0], expectedClusterResponse)

	cluster, err := NewGKECluster(client, "hello", "wassup", "hola", int32(0))
	if err != nil {
		t.Errorf("error creating cluster handle: %s", err)
	}

	mockClusterManager.resps = append(mockClusterManager.resps[:0], expectedResponse)

	resp, err := cluster.LatestMasterVersionForReleaseSeries(context.Background(), "1.10")

	if err != nil {
		t.Fatal(err)
	}

	if want, got := expectedResponse.ValidMasterVersions[0], resp; want != got {
		t.Errorf("wrong response %q, want %q)", got, want)
	}

	resp, err = cluster.LatestMasterVersionForReleaseSeries(context.Background(), "1.9")

	if err != nil {
		t.Fatal(err)
	}

	if want, got := expectedResponse.ValidMasterVersions[1], resp; want != got {
		t.Errorf("wrong response %q, want %q)", got, want)
	}

	resp, err = cluster.LatestMasterVersionForReleaseSeries(context.Background(), "1.9.1")

	if err != nil {
		t.Fatal(err)
	}

	if want, got := expectedResponse.ValidMasterVersions[3], resp; want != got {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestLatestNodeVersionForReleaseSeries(t *testing.T) {
	var validNodeVersions = []string{
		"1.10.1-gke.5",
		"1.9.2-gke.1",
		"1.9.2-gke.0",
		"1.9.1-gke.0",
		"1.8.4-gke.1",
	}
	var expectedResponse = &containerpb.ServerConfig{
		ValidNodeVersions: validNodeVersions,
	}

	var expectedClusterResponse = &containerpb.Cluster{
		CurrentMasterVersion: "1.9.2-gke.1",
	}

	client, _ := container.NewClusterManagerClient(context.Background(), clientOpt)

	mockClusterManager.err = nil
	mockClusterManager.reqs = nil

	mockClusterManager.resps = append(mockClusterManager.resps[:0], expectedClusterResponse)

	cluster, err := NewGKECluster(client, "hello", "wassup", "hola", int32(0))
	if err != nil {
		t.Errorf("error creating cluster handle: %s", err)
	}

	mockClusterManager.resps = append(mockClusterManager.resps[:0], expectedResponse)

	resp, err := cluster.LatestNodeVersionForReleaseSeries(context.Background(), "1.11")

	if err == nil {
		t.Errorf("expected error was not thrown, got %s", resp)
	}

	resp, err = cluster.LatestNodeVersionForReleaseSeries(context.Background(), "1.10")

	if err == nil {
		t.Errorf("expected error was not thrown, got %s", resp)
	}

	resp, err = cluster.LatestNodeVersionForReleaseSeries(context.Background(), "1.9")

	if err != nil {
		t.Error(err)
	}

	if want, got := expectedResponse.ValidNodeVersions[1], resp; want != got {
		t.Errorf("wrong response %q, want %q)", got, want)
	}

	resp, err = cluster.LatestNodeVersionForReleaseSeries(context.Background(), "1.9.1")

	if err != nil {
		t.Error(err)
	}

	if want, got := expectedResponse.ValidNodeVersions[3], resp; want != got {
		t.Errorf("wrong response %q, want %q)", got, want)
	}

	resp, err = cluster.LatestNodeVersionForReleaseSeries(context.Background(), "1.8")

	if err != nil {
		t.Error(err)
	}

	if want, got := expectedResponse.ValidNodeVersions[4], resp; want != got {
		t.Errorf("wrong response %q, want %q)", got, want)
	}

	resp, err = cluster.LatestNodeVersionForReleaseSeries(context.Background(), "1.7")

	if err == nil {
		t.Errorf("expected error was not thrown, got %s", resp)
	}
}
