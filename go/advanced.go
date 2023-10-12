package main

import (
	"context"

	"github.com/nexus-rpc/sdk-go/nexus"
)

var startWorkflowWithMapperOp = nexus.WithMapper(
	startWorkflowOp,
	func(ctx context.Context, mo *CreateCellOutput, uoe *nexus.UnsuccessfulOperationError) (MyMappedOutput, *nexus.UnsuccessfulOperationError, error) {
		return MyMappedOutput{}, nil, nil
	},
)

// type tenantIDKey struct{}

// func constructID(ctx context.Context, operation string, parts ...string) string {
// 	// tenantID := ctx.Value(tenantIDKey{}).(string)

// 	// return operation + "-" + tenantID + "-" + strings.Join(parts, "-")
// 	return operation + "-" + strings.Join(parts, "-")
// }

// func hashSource(s string) string {
// 	sum := sha1.Sum([]byte(s))
// 	return fmt.Sprintf("%x", sum[:8])
// }

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
