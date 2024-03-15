package api

import (
	"context"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
)

const ServiceName = "infra"

type CreateCellInput struct {
	CellID    string
	Nexusness int64
}

type CreateCellOutput struct {
	CellID string
}

var StartWorkflowOp1 = NewWorkflowRunOperation("provision-cell", MyCalleeWorkflow,
	func(ctx context.Context, input CreateCellInput) (client.StartWorkflowOptions, error) {
		options := client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}
		return options, nil
	},
)

// ----------------------------------------------------------
// or optionally with an op ref that can be used separately
// ----------------------------------------------------------
var StartWorkflowOpRef = OperationRef[CreateCellInput, *CreateCellOutput]("provision-cell")

var StartWorkflowOp2 = NewWorkflowRunOperation(StartWorkflowOpRef, MyCalleeWorkflow,
	func(ctx context.Context, input CreateCellInput) (client.StartWorkflowOptions, error) {
		options := client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}
		return options, nil
	},
)

// ----------------------------------------------------------
// or optionally with mapping workflow IO types to op IO types
// ----------------------------------------------------------

var StartWorkflowOp3 = NewWorkflowRunOperationWithMapping[CreateCellInput, *CreateCellOutput]("provision-cell", MyCalleeWorkflowDifferentTypes,
	func(ctx context.Context, input CreateCellInput) (client.StartWorkflowOptions, CreateCellWorkflowInput, error) {
		options := client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}
		wfInput := CreateCellWorkflowInput{ID: input.CellID}
		return options, wfInput, nil
	},
	//TODO: add output mapping, which will also give us the types for inference, for use with a string op name "provision-cell" as the 1st arg
)

var StartWorkflowOp = NewWorkflowRunOperationWithMapping(StartWorkflowOpRef, MyCalleeWorkflowDifferentTypes,
	func(ctx context.Context, input CreateCellInput) (client.StartWorkflowOptions, CreateCellWorkflowInput, error) {
		options := client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}
		wfInput := CreateCellWorkflowInput{ID: input.CellID}
		return options, wfInput, nil
	},
	//TODO: add output mapping, which will also give us the types for inference, for use with a string op name "provision-cell" as the 1st arg
)

var QueryOp = temporalnexus.NewSyncOperation("get-cell-status", func(ctx context.Context, c client.Client, cellID string) (string, error) {
	payload, err := c.QueryWorkflow(ctx, "provision-cell-"+cellID, "", "get-cell-status")
	if err != nil {
		return "", err
	}
	var status string
	return status, payload.Get(&status)
})

var SignalOp = temporalnexus.NewSyncOperation("resume-provisioning", func(ctx context.Context, c client.Client, cellID string) (nexus.Void, error) {
	// TODO: signal request should use Nexus request ID
	return nil, c.SignalWorkflow(ctx, "provision-cell-"+cellID, "", "resume", nil)
})
