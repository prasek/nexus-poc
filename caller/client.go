package caller

import (
	"github.com/bergundy/nexus-poc/callee/api"
	"go.temporal.io/sdk/workflow"
)

type CalleeWorkflowClientItf interface {
	StartProvisionCell(input api.CreateCellInput) (HandleProvisionCell, error)
	GetCellStatus(cellID string) (string, error)
	ResumeProvisioning(cellID string) error
}

type HandleProvisionCell workflow.OperationHandle[*api.CreateCellOutput]

type CalleeWorkflowClient struct {
	ctx workflow.Context
}

func NewCalleeWorkflowClient(ctx workflow.Context) CalleeWorkflowClientItf {
	return &CalleeWorkflowClient{ctx: ctx}
}

func (client *CalleeWorkflowClient) StartProvisionCell(input api.CreateCellInput) (HandleProvisionCell, error) {
	startHandle, err := workflow.StartOperation(client.ctx, api.ServiceName, api.StartWorkflowOp, input, workflow.OperationOptions{})
	if err != nil {
		return nil, err
	}
	if err = startHandle.WaitStarted(client.ctx); err != nil {
		return nil, err
	}
	return startHandle, nil
}

func (client *CalleeWorkflowClient) GetCellStatus(cellID string) (string, error) {
	queryHandle, err := workflow.StartOperation(client.ctx, api.ServiceName, api.QueryOp, cellID, workflow.OperationOptions{})
	if err != nil {
		return "", err
	}
	qOut, err := queryHandle.GetResult(client.ctx)
	if err != nil {
		return "", err
	}
	return qOut, nil
}

func (client *CalleeWorkflowClient) ResumeProvisioning(cellID string) error {
	signalHandle, err := workflow.StartOperation(client.ctx, api.ServiceName, api.SignalOp, cellID, workflow.OperationOptions{})
	if err != nil {
		return err
	}
	if _, err := signalHandle.GetResult(client.ctx); err != nil {
		return err
	}
	return nil
}
