package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bergundy/nexus-poc/callee/api"
	"github.com/bergundy/nexus-poc/caller"
	"github.com/bergundy/nexus-poc/setup"
	"go.temporal.io/sdk/client"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	opts := setup.GetOptions(os.Args[1:])
	if !opts.SkipEnvSetup {
		setup.SetupEnv(ctx, opts, api.ServiceName)
	}

	callerClient, _ := setup.CreateClients(opts)

	run, err := callerClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		TaskQueue: "my-caller-queue",
	}, caller.MyCallerWorkflow)
	if err != nil {
		log.Panic(err)
	}
	var out api.CreateCellOutput
	err = run.Get(context.Background(), &out)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("done:", out)
}
