package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type MyInput struct {
	CellID string
	Stuff  int64
}
type MyOutput struct {
}
type MyIntermediateOutput struct{}
type MyMappedOutput struct{}

func MyHandlerWorkflow(ctx workflow.Context, input MyInput) (MyOutput, error) {
	return MyOutput{}, nil
}

func MyCallerWorkflow(ctx workflow.Context) (MyOutput, error) {
	// handle, _ := workflow.StartOperation(ctx, startWorkflowOp, MyInput{})
	// _ = handle.WaitStarted(ctx)
	// _, _ = workflow.StartOperation(ctx, startWorkflowWithMapperOp, MyInput{})
	handle, err := workflow.StartOperation(ctx, queryOp, MyInput{})
	if err != nil {
		return MyOutput{}, nil
	}
	return handle.GetResult(ctx)
}

var startWorkflowSimple = temporalnexus.NewWorkflowRunOperation("provision-cell-simple", temporalnexus.WorkflowRunOptions[MyInput, MyOutput]{
	Workflow: MyHandlerWorkflow,
	GetOptions: func(ctx context.Context, input MyInput) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID: constructID(ctx, "provision-cell", input.CellID),
		}, nil
	},
})

var startWorkflowOp = temporalnexus.NewWorkflowRunOperation("provision-cell", temporalnexus.WorkflowRunOptions[MyInput, MyOutput]{
	Start: func(ctx context.Context, c client.Client, input MyInput) (temporalnexus.WorkflowHandle[MyOutput], error) {
		fmt.Println("Starting workflow")
		return temporalnexus.StartWorkflow(ctx, c, client.StartWorkflowOptions{
			ID: constructID(ctx, "provision-cell", input.CellID),
		}, MyHandlerWorkflow, input)
	},
})

var queryOp = temporalnexus.NewSyncOperation("get-cell-status", func(ctx context.Context, c client.Client, input MyInput) (MyOutput, error) {
	fmt.Println("Callback called!")
	return MyOutput{}, nil
	// payload, _ := c.QueryWorkflow(ctx, constructID(ctx, "provision-cell", input.CellID), "", "get-cell-status")
	// var output MyOutput
	// return output, payload.Get(&output)
})

var signalOp = temporalnexus.NewSyncOperation("set-cell-status", func(ctx context.Context, c client.Client, input MyInput) (nexus.NoResult, error) {
	return nil, c.SignalWorkflow(ctx, constructID(ctx, "provision-cell", input.CellID), "", "set-cell-status", input)
})

var startWorkflowWithMapperOp = nexus.WithMapper[MyInput, MyOutput, MyOutput, MyMappedOutput](
	startWorkflowOp,
	func(ctx context.Context, mo MyOutput, uoe *nexus.UnsuccessfulOperationError) (MyMappedOutput, *nexus.UnsuccessfulOperationError, error) {
		return MyMappedOutput{}, nil, nil
	},
)

func main() {
	ctx := context.TODO()
	c, err := client.Dial(client.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
	})
	if err != nil {
		log.Panic(err)
	}
	rp := time.Hour * 24
	c.WorkflowService().RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        "default",
		WorkflowExecutionRetentionPeriod: &rp,
	})
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	w := worker.New(c, "my-task-queue", worker.Options{
		NexusOperations: []nexus.UntypedOperationHandler{
			startWorkflowSimple,
			startWorkflowWithMapperOp,
			queryOp,
			signalOp,
		},
	})
	w.RegisterWorkflow(MyCallerWorkflow)
	w.RegisterWorkflow(MyHandlerWorkflow)
	w.RegisterActivityWithOptions(func(ctx context.Context, input any) (any, error) {
		nc, err := nexus.NewClient(nexus.ClientOptions{ServiceBaseURL: "http://localhost:7253/foo"})
		if err != nil {
			return nil, err
		}
		codec := nexus.JSONCodec{}
		message, err := codec.ToMessage(input)
		if err != nil {
			return nil, err
		}
		u, err := url.Parse("http://localhost:7253/system/callback")
		ai := activity.GetInfo(ctx)
		q := u.Query()
		q.Add("workflow_id", ai.WorkflowExecution.ID)
		// TODO: get the ID
		q.Add("signal_name", fmt.Sprintf("operation-%v", 1))
		u.RawQuery = q.Encode()
		if err != nil {
			return nil, err
		}
		response, err := nc.StartOperation(context.TODO(), nexus.StartOperationOptions{
			Operation:   startWorkflowOp.Name,
			Header:      message.Header,
			Body:        message.Body,
			CallbackURL: u.String(),
		})
		if err != nil {
			return nil, err
		}
		// TODO: need some way to indicate which response type we got
		if response.Successful != nil {
			defer response.Successful.Body.Close()
			var out any
			err = codec.FromMessage(&nexus.Message{Header: response.Successful.Header, Body: response.Successful.Body}, &out)
			return out, err
		}
		return response.Pending.ID, nil

	}, activity.RegisterOptions{
		Name: "start-operation",
	})
	w.Start()
	defer w.Stop()

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		TaskQueue: "my-task-queue",
	}, MyCallerWorkflow)
	if err != nil {
		log.Panic(err)
	}
	var out MyOutput
	err = run.Get(context.Background(), &out)
	if err != nil {
		log.Panic(err)
	}
}

type tenantIDKey struct{}

func constructID(ctx context.Context, operation string, parts ...string) string {
	// tenantID := ctx.Value(tenantIDKey{}).(string)

	// return operation + "-" + tenantID + "-" + strings.Join(parts, "-")
	return operation + "-" + strings.Join(parts, "-")
}

func hashSource(s string) string {
	sum := sha1.Sum([]byte(s))
	return fmt.Sprintf("%x", sum[:8])
}

// type operationKey struct{}

// // w := worker.New(c, "my-task-queue", worker.Options{Interceptors: []interceptor.WorkerInterceptor{&OperationAuthorizationInterceptor{}, &ReencryptionInterceptor{}}})
// type OperationAuthorizationInterceptor struct {
// 	interceptor.WorkerInterceptorBase
// 	interceptor.OperationInboundInterceptorBase
// 	interceptor.OperationOutboundInterceptorBase
// }

// func (i *OperationAuthorizationInterceptor) getTenantID(h http.Header) (string, error) {
// 	source := h.Get("Temporal-Source-Namespace")
// 	if source == "" {
// 		return "", operation.NewUnauthorizedError("unauthorized access")
// 	}
// 	return hashSource(source), nil
// }

// func (i *OperationAuthorizationInterceptor) StartOperation(ctx context.Context, request *nexus.StartOperationRequest) (nexus.OperationResponse, error) {
// 	tenantID, err := i.getTenantID(request.HTTPRequest.Header)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ctx = context.WithValue(ctx, tenantIDKey{}, tenantID)
// 	ctx = context.WithValue(ctx, operationKey{}, request.Operation)
// 	return i.OperationInboundInterceptorBase.Next.StartOperation(ctx, request)
// }

// func (i *OperationAuthorizationInterceptor) CancelOperation(ctx context.Context, request *nexus.CancelOperationRequest) error {
// 	tenantID, err := i.getTenantID(request.HTTPRequest.Header)
// 	if err != nil {
// 		return err
// 	}

// 	if !strings.HasPrefix(request.OperationID, fmt.Sprintf("%s:%s:", request.Operation, tenantID)) {
// 		return operation.NewUnauthorizedError("unauthorized access")
// 	}
// 	return i.OperationInboundInterceptorBase.Next.CancelOperation(ctx, request)
// }

// func (i *OperationAuthorizationInterceptor) ExecuteOperation(ctx context.Context, input *interceptor.ClientExecuteWorkflowInput) (client.WorkflowRun, error) {
// 	operation := ctx.Value(operationKey{}).(string)

// 	if !strings.HasPrefix(input.Options.ID, constructID(ctx, operation)) {
// 		return nil, errors.New("Workflow ID does not match expected format")
// 	}

// 	return i.OperationOutboundInterceptorBase.Next.ExecuteOperation(ctx, input)
// }

// type encryptionKeyKey struct{}

// type ReencryptionInterceptor struct {
// 	interceptor.WorkerInterceptorBase
// 	interceptor.OperationInboundInterceptorBase
// 	interceptor.OperationOutboundInterceptorBase
// }

// func (i *ReencryptionInterceptor) StartOperation(ctx context.Context, request *nexus.StartOperationRequest) (nexus.OperationResponse, error) {
// 	responseEncryptionKey := request.HTTPRequest.Header.Get("Response-Encryption-Key")
// 	if responseEncryptionKey != "" {
// 		ctx = context.WithValue(ctx, encryptionKeyKey{}, responseEncryptionKey)
// 	}

// 	return i.OperationInboundInterceptorBase.Next.StartOperation(ctx, request)
// }

// func (i *ReencryptionInterceptor) ExecuteOperation(ctx context.Context, input *interceptor.ClientExecuteWorkflowInput) (client.WorkflowRun, error) {
// 	if _, ok := ctx.Value(encryptionKeyKey{}).(string); ok {
// 		opts := *input.Options
// 		input.Options = &opts
// 		// TODO: inject the key into the callback context.
// 		return i.OperationOutboundInterceptorBase.Next.ExecuteOperation(ctx, input)
// 	}

// 	return i.OperationOutboundInterceptorBase.Next.ExecuteOperation(ctx, input)
// }

// func (i *ReencryptionInterceptor) MapCompletion(ctx context.Context, request *MapCompletionRequest) (nexus.OperationCompletion, error) {
// 	// converter.NewEnc
// }
