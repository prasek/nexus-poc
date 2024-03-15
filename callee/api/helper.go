package api

import (
	"context"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"
)

type OperationReference[I, O any] interface {
	Name() string
	// A type inference helper for implementations of this interface.
	inferType(I, O)
}

type operationReference[I, O any] string

func (r operationReference[I, O]) Name() string {
	return string(r)
}

func (operationReference[I, O]) inferType(I, O) {} //nolint:unused

// NewOperationReference creates an [OperationReference] with the provided type parameters and name.
// It provides typed interface for invoking operations when the implementation is not available to the caller.
func NewOperationReference[I, O any](name string) OperationReference[I, O] {
	return operationReference[I, O](name)
}

type WorkflowOpHandlerFunc[I, WFI any] func(input I) (WFI, client.StartWorkflowOptions)

func NewWorkflowRunOperation[I, O, WFI, WFO any](ref OperationReference[I, O], wf func(ctx workflow.Context, input WFI) (WFO, error), handler WorkflowOpHandlerFunc[I, WFI]) *nexus.AsyncOperation[I, O, O] {
	return temporalnexus.NewWorkflowRunOperation[I, O](ref.Name(), temporalnexus.WorkflowRunOptions[I, O]{
		Start: func(ctx context.Context, c client.Client, input I) (temporalnexus.WorkflowHandle[O], error) {
			wfInput, options := handler(input)
			return temporalnexus.StartWorkflow(ctx, c, options, wf, wfInput)
		},
	})
}
