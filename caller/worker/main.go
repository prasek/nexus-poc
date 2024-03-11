package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bergundy/nexus-poc/callee/api"
	"github.com/bergundy/nexus-poc/caller"
	"github.com/bergundy/nexus-poc/setup"
	"go.temporal.io/sdk/worker"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	opts := setup.GetOptions(os.Args[1:])
	if !opts.SkipEnvSetup {
		setup.SetupEnv(ctx, opts, api.ServiceName)
	}

	callerClient, _ := setup.CreateClients(opts)

	w := worker.New(callerClient, "my-caller-queue", worker.Options{})
	w.RegisterWorkflow(caller.MyCallerWorkflow)
	err := w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Unable to start caller worker")
	}
}
