package api

import (
	"go.temporal.io/sdk/workflow"
)

type CreateCellWorkflowInput struct {
	ID         string
	Nexusness  int64
	MoreThings any
}

/*
type CreateCellWorkflowOutput struct {
	ID         string
	MoreThings any
}
*/

func MyCalleeWorkflow(ctx workflow.Context, input CreateCellInput) (*CreateCellOutput, error) {
	workflow.SetQueryHandler(ctx, "get-cell-status", func() (string, error) {
		return "running", nil
	})
	ch := workflow.GetSignalChannel(ctx, "resume")
	ch.Receive(ctx, nil)

	//WF has knowledge of Op return type
	return &CreateCellOutput{CellID: input.CellID}, nil
}

func MyCalleeWorkflowDifferentTypes(ctx workflow.Context, input CreateCellWorkflowInput) (*CreateCellOutput, error) {
	workflow.SetQueryHandler(ctx, "get-cell-status", func() (string, error) {
		return "running", nil
	})
	ch := workflow.GetSignalChannel(ctx, "resume")
	ch.Receive(ctx, nil)

	//WF has knowledge of Op return type
	return &CreateCellOutput{CellID: input.ID}, nil
}
