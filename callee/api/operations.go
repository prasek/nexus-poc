package api

import (
	"context"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
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

var StartWorkflowOpRef = NewOperationReference[CreateCellInput, *CreateCellOutput]("provision-cell")

var StartWorkflowOp = NewWorkflowRunOperation(
	StartWorkflowOpRef,
	MyCalleeWorkflow,
	func(input CreateCellInput) (CreateCellWorkflowInput, client.StartWorkflowOptions) {
		wfInput := CreateCellWorkflowInput{
			ID: input.CellID,
		}

		options := client.StartWorkflowOptions{
			ID:                       "provision-cell-" + input.CellID,
			WorkflowExecutionTimeout: time.Second * 10,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second * 1,
				BackoffCoefficient: 2,
				MaximumInterval:    time.Second * 60,
			},
		}

		return wfInput, options

	})

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

/*
// note: in this POC the Nexus op output must match the workflow output - in the future we'll allow mapping the workflow output
var StartWorkflowOp2 = temporalnexus.NewWorkflowRunOperation("provision-cell", temporalnexus.WorkflowRunOptions[CreateCellInput, *CreateCellOutput]{
	Start: func(ctx context.Context, c client.Client, input CreateCellInput) (temporalnexus.WorkflowHandle[*CreateCellOutput], error) {
		wfInput := CreateCellWorkflowInput{
			ID: input.CellID,
		}

		return temporalnexus.StartWorkflow(ctx, c, client.StartWorkflowOptions{
			ID:                       "provision-cell-" + input.CellID,
			WorkflowExecutionTimeout: time.Second * 10,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Second * 1,
				BackoffCoefficient: 2,
				MaximumInterval:    time.Second * 60,
			},
		}, MyCalleeWorkflow, wfInput)
	},
})
*/

/*
var StartSimpleWorkflowOp = temporalnexus.NewWorkflowRunOperation("provision-cell", temporalnexus.WorkflowRunOptions[CreateCellInput, *CreateCellOutput]{
	//if using the simpler `Workflow` option, the I,O of the nexus op and the workflow must match
	//cannot use (func(ctx context.Context, c client.Client, input CreateCellInput) (temporalnexus.WorkflowHandle[*CreateCellOutput], error) literal) (value of type func(ctx context.Context, c client.Client, input CreateCellInput) (temporalnexus.WorkflowHandle[*CreateCellOutput], error)) as func(context.Context, client.Client, CreateCellWorkflowInput) (temporalnexus.WorkflowHandle[*CreateCellOutput], error) value in struct literalcompilerIncompatibleAssign
	Workflow: MyCalleeWorkflow,

	GetOptions: func(ctx context.Context, input CreateCellInput) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID: "provision-cell-" + input.CellID,
		}, nil
	},
})
*/
