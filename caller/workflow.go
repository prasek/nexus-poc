package caller

import (
	"github.com/bergundy/nexus-poc/callee/api"
	"go.temporal.io/sdk/workflow"
)

func MyCallerWorkflow(ctx workflow.Context) (*api.CreateCellOutput, error) {
	cellID := "s-nexus"
	input := api.CreateCellInput{CellID: cellID, Nexusness: 100}
	startHandle, err := workflow.StartOperation(ctx, api.ServiceName, api.StartWorkflowOp, input, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	if err = startHandle.WaitStarted(ctx); err != nil {
		return nil, err
	}
	queryHandle, err := workflow.StartOperation(ctx, api.ServiceName, api.QueryOp, cellID, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	qOut, err := queryHandle.GetResult(ctx)
	if err != nil {
		return nil, err
	}
	workflow.GetLogger(ctx).Info("got cell status", "status", qOut)
	signalHandle, err := workflow.StartOperation(ctx, api.ServiceName, api.SignalOp, cellID, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	if _, err := signalHandle.GetResult(ctx); err != nil {
		return nil, err
	}
	return startHandle.GetResult(ctx)
}
