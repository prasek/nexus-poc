package main

import (
	"context"
	"log"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const serviceName = "infra"

type CreateCellInput struct {
	CellID    string
	Nexusness int64
}

type CreateCellOutput struct {
	CellID string
}

func MyHandlerWorkflow(ctx workflow.Context, input CreateCellInput) (*CreateCellOutput, error) {
	workflow.SetQueryHandler(ctx, "get-cell-status", func() (string, error) {
		return "running", nil
	})
	ch := workflow.GetSignalChannel(ctx, "resume")
	ch.Receive(ctx, nil)
	return &CreateCellOutput{CellID: input.CellID}, nil
}

func MyCallerWorkflow(ctx workflow.Context) (*CreateCellOutput, error) {
	cellID := "s-nexus"
	input := CreateCellInput{CellID: cellID, Nexusness: 100}
	startHandle, err := workflow.StartOperation(ctx, serviceName, startWorkflowOp, input, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	if err = startHandle.WaitStarted(ctx); err != nil {
		return nil, err
	}
	queryHandle, err := workflow.StartOperation(ctx, serviceName, queryOp, cellID, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	qOut, err := queryHandle.GetResult(ctx)
	if err != nil {
		return nil, err
	}
	workflow.GetLogger(ctx).Info("got cell status", "status", qOut)
	signalHandle, err := workflow.StartOperation(ctx, serviceName, signalOp, cellID, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	if _, err := signalHandle.GetResult(ctx); err != nil {
		return nil, err
	}
	return startHandle.GetResult(ctx)
}

// Alternative 1
var startWorkflowOp = temporalnexus.NewWorkflowRunOperation("provision-cell", temporalnexus.WorkflowRunOptions[CreateCellInput, *CreateCellOutput]{
	Start: func(ctx context.Context, c client.Client, input CreateCellInput) (temporalnexus.WorkflowHandle[*CreateCellOutput], error) {
		return temporalnexus.StartWorkflow(ctx, c, client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}, MyHandlerWorkflow, input)
	},
})

// Alternative 2
var startWorkflowSimple = temporalnexus.NewWorkflowRunOperation("provision-cell", temporalnexus.WorkflowRunOptions[CreateCellInput, *CreateCellOutput]{
	Workflow: MyHandlerWorkflow,
	GetOptions: func(ctx context.Context, input CreateCellInput) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}, nil
	},
})

var queryOp = temporalnexus.NewSyncOperation("get-cell-status", func(ctx context.Context, c client.Client, cellID string) (string, error) {
	payload, err := c.QueryWorkflow(ctx, "provision-cell-"+cellID, "", "get-cell-status")
	if err != nil {
		return "", err
	}
	var status string
	return status, payload.Get(&status)
})

var signalOp = temporalnexus.NewSyncOperation("resume-provisioning", func(ctx context.Context, c client.Client, cellID string) (nexus.Void, error) {
	return nil, c.SignalWorkflow(ctx, "provision-cell-"+cellID, "", "resume", nil)
})

func setup(ctx context.Context) (client.Client, client.Client) {
	callerClient, err := client.Dial(client.Options{
		HostPort:  "localhost:7233",
		Namespace: "caller",
	})
	if err != nil {
		log.Panic(err)
	}

	handlerClient, err := client.Dial(client.Options{
		HostPort:  "localhost:7233",
		Namespace: "handler",
	})
	if err != nil {
		log.Panic(err)
	}

	retention := time.Hour * 24
	callerClient.WorkflowService().RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        "caller",
		WorkflowExecutionRetentionPeriod: &retention,
	})
	callerClient.WorkflowService().RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        "handler",
		WorkflowExecutionRetentionPeriod: &retention,
	})

	_, err = callerClient.WorkflowService().UpdateNamespace(ctx, &workflowservice.UpdateNamespaceRequest{
		Namespace: "caller",
		UpdateInfo: &namespace.UpdateNamespaceInfo{
			OutgoingServiceUpdates: []*namespace.OutgoingServiceUpdate{
				{Variant: &namespace.OutgoingServiceUpdate_CreateOrUpdateService_{
					CreateOrUpdateService: &namespace.OutgoingServiceUpdate_CreateOrUpdateService{
						Name:    serviceName,
						BaseUrl: "http://localhost:7253/" + serviceName,
					},
				}},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = callerClient.OperatorService().CreateOrUpdateNexusIncomingService(ctx, &operatorservice.CreateOrUpdateNexusIncomingServiceRequest{
		NexusIncomingService: &operatorservice.NexusIncomingService{
			Name:      serviceName,
			Namespace: "handler",
			TaskQueue: "my-handler-queue",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return callerClient, handlerClient
}

func startHandler(c client.Client) worker.Worker {
	w := worker.New(c, "my-handler-queue", worker.Options{
		NexusOperations: []nexus.UntypedOperationHandler{
			// startWorkflowOp,
			startWorkflowSimple,
			queryOp,
			signalOp,
		},
	})
	w.RegisterWorkflow(MyHandlerWorkflow)
	w.Start()
	return w
}

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	callerClient, handlerClient := setup(ctx)
	handlerWorker := startHandler(handlerClient)
	defer handlerWorker.Stop()

	w := worker.New(callerClient, "my-caller-queue", worker.Options{})
	w.RegisterWorkflow(MyCallerWorkflow)
	w.Start()
	defer w.Stop()

	run, err := callerClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		TaskQueue: "my-caller-queue",
	}, MyCallerWorkflow)
	if err != nil {
		log.Panic(err)
	}
	var out CreateCellOutput
	err = run.Get(context.Background(), &out)
	if err != nil {
		log.Panic(err)
	}
}
