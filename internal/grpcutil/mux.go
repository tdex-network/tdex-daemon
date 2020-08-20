package grpcutil

import (
	"net"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

// ServeMux serves a given grpc server and multiplexes on the same tcp port a grpc-web server
// The grpc request should have the content-type application/grpc-web to be routed over Grpc Web
func ServeMux(address string, grpcServer *grpc.Server) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := mux.Match(cmux.HTTP1Fast())

	grpcWebServer := grpcweb.WrapServer(grpcServer)

	go grpcServer.Serve(grpcL)
	go http.Serve(httpL, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if grpcWebServer.IsGrpcWebRequest(req) {
			grpcWebServer.ServeHTTP(resp, req)
		}
	}))

	go mux.Serve()
	return nil
}
