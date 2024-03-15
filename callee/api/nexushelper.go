package api

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"
)

type OperationRef[I, O any] string

func (s OperationRef[I, O]) Name() string   { return string(s) }
func (s OperationRef[I, O]) inferType(I, O) {} //nolint:unused

func NewWorkflowRunOperation[I, O any](
	name OperationRef[I, O],
	workflow func(workflow.Context, I) (O, error),
	f func(I) client.StartWorkflowOptions,
) *nexus.AsyncOperation[I, O, O] {
	fmt.Printf("Name: %s", string(name))
	return temporalnexus.NewWorkflowRunOperation[I, O](string(name), temporalnexus.WorkflowRunOptions[I, O]{
		Start: func(ctx context.Context, c client.Client, input I) (temporalnexus.WorkflowHandle[O], error) {
			options := f(input)
			return temporalnexus.StartWorkflow(ctx, c, options, workflow, input)
		},
	})
}

func NewWorkflowRunOperationWithMapping[I, O, WFI, WFO any](
	name OperationRef[I, O],
	workflow func(workflow.Context, WFI) (WFO, error),
	f func(I) (client.StartWorkflowOptions, WFI),
) *nexus.AsyncOperation[I, O, O] {
	return temporalnexus.NewWorkflowRunOperation[I, O](string(name), temporalnexus.WorkflowRunOptions[I, O]{
		Start: func(ctx context.Context, c client.Client, input I) (temporalnexus.WorkflowHandle[O], error) {
			options, wfInput := f(input)
			return temporalnexus.StartWorkflow(ctx, c, options, workflow, wfInput)
		},
	})
}
