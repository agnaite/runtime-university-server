# runtime-university-server
A gRPC server used as part of the [Runtime University coding project](https://salesforce.quip.com/0kszArgc4unE)

This repo builds on the work down on [hello-grpc](https://github.com/heroku/hello-grpc) to run a
gRPC server in a private space. The server implements the [route_guide](https://github.com/grpc/grpc-go/tree/master/examples/route_guide)
gRPC server for the purpose of having Runtime University members write a client that connects to the service.

Currently, this is deployed to the `runtime-university` private space in an app named `grpc-server`.
