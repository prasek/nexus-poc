package caller

import (
	"github.com/bergundy/nexus-poc/callee/api"
	"go.temporal.io/sdk/workflow"
)

func MyCallerWorkflow(ctx workflow.Context) (*api.CreateCellOutput, error) {
	callee := NewCalleeWorkflowClient(ctx)

	input := api.CreateCellInput{CellID: "r-nexus", Nexusness: 100}
	startHandle, err := callee.StartProvisionCell(input)
	if err != nil {
		return nil, err
	}

	status, err := callee.GetCellStatus(input.CellID)
	if err != nil {
		return nil, err
	}

	workflow.GetLogger(ctx).Info("got cell status", "status", status)

	err = callee.ResumeProvisioning(input.CellID)
	if err != nil {
		return nil, err
	}

	return startHandle.GetResult(ctx)
}
