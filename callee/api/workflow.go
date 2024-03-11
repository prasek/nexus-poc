package api

import (
	"go.temporal.io/sdk/workflow"
)

func MyCalleeWorkflow(ctx workflow.Context, input CreateCellInput) (*CreateCellOutput, error) {
	workflow.SetQueryHandler(ctx, "get-cell-status", func() (string, error) {
		return "running", nil
	})
	ch := workflow.GetSignalChannel(ctx, "resume")
	ch.Receive(ctx, nil)
	return &CreateCellOutput{CellID: input.CellID}, nil
}
