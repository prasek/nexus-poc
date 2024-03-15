package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bergundy/nexus-poc/callee/api"
	"github.com/bergundy/nexus-poc/setup"
	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/worker"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	opts := setup.GetOptions(os.Args[1:])
	if !opts.SkipEnvSetup {
		setup.SetupEnv(ctx, opts, api.ServiceName)
	}

	_, handlerClient := setup.CreateClients(opts)

	w := worker.New(handlerClient, "my-handler-queue", worker.Options{
		NexusOperations: []nexus.UntypedOperationHandler{
			api.StartWorkflowOp,
			api.QueryOp,
			api.SignalOp,
		},
	})
	w.RegisterWorkflow(api.MyCalleeWorkflow)
	w.RegisterWorkflow(api.MyCalleeWorkflowDifferentTypes)
	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start callee worker")
	}
}
