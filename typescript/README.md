## Hand coded Nexus handler

### Lowest handler level

#### Sync request

`worker.ts`

```ts
import { Worker } from "@temporalio/worker";
import { ApplicationFailure } from "@temporalio/common";
import { Request, Response } from "@temporalio/nexus"; // name TBD

await Worker.create({
  workflowsPath: require.resolve("./workflows"),
  activities,
  async operationHandler({
    service,
    operation,
    headers,
    body,
  }: Request): Promise<Response> {
    // handler is essentially a router.
    if (service === "calculator" && operation === "add") {
      if (!headers["content-type"].startsWith("application/json")) {
        // Nexus has a notion of retryable and non-retryable errors.
        // Non-`ApplicationFailure` errors thrown in a handler are converted to retryable `ApplicationFailure`s.
        // TODO: Decide if we want to reuse the ApplicationFailure class here.
        throw ApplicationFailure.nonRetryable("Invalid content type");
      }

      const { x, y } = JSON.parse(body);
      const result = x + y;

      return {
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ result }),
      };
    }
  },
});
```

#### Start operation

```ts
async function operationHandler({
  service,
  operation,
  headers,
  body,
  callbackURL,
}: Request): Promise<Response> {
  // TODO(me): Are these the right header and status code?
  if (headers["authorization"] !== "Bearer top-secret") {
    // Later we may add an HTTP gateway in front of Temporal (At a later stage in the project).
    // When we get there, we'll need to support setting the status code (as well as other HTTP specific features) in handlers.
    // HTTP specific data may be delivered in specific response headers.
    throw new ApplicationFailure.nonRetryable("Unauthorized");
  }

  if (service === "payments" && operation === "checkout") {
    if (!headers["content-type"].startsWith("application/json")) {
      // Nexus has a notion of retryable and non-retryable errors
      throw ApplicationFailure.nonRetryable("Invalid content type");
    }

    const customerId = await getCustomerInfo(headers);

    const { account, requestId, product } = JSON.parse(body);
    const workflowId = `${customerId}-${account}-${requestId}`;

    const handle = await client.workflow.start(workflows.payment, {
      workflowId,
      args: [{ account, requestId, amount }],
      callbackURL,
    });
    return {
      body: JSON.stringify({ result }),
    };
  }
}
```

#### Data conversion

```ts
await Worker.create({
  workflowsPath: require.resolve("./workflows"),
  activities,
  async operationHandler(request: Request): Promise<Response> {
    const { service, operation, headers, body } = request;
    // This seems redunant in the Temporal context but handler is Temporal agnostic.
    const payload = { data: body, metadata: headers };

    // handler is essentially a router.
    if (service === "my-service" && operation === "add") {
      const input: number = await myDataConverter.fromPayload(payload);
      const output = input + 1;
      const outputPayload = await myDataConverter.toPayload(output);

      return {
        headers: { "Content-Type": "application/json" },
        // OR
        // headers is a map of string to string
        // metadata accepts byte arrays, conversion required
        headers: payloadMetadataToHeaders(outputPayload.metadata),
        body: outputEncoded,
      };

      // OR with a helper:
      return {
        payload: outputPayload,
      };
    }
  },
});
```

<!--  -->
<!-- // export async function myWorkflow() { -->
<!-- // } -->
<!--  -->
<!-- // workflows.ts -->
<!-- // Option 1: register in workflow context -->
<!-- export const myOperation = workflow.defineOperation({ -->
<!--   name: 'my-operation', -->
<!--   cancellable: true, -->
<!--   async handler() {}, -->
<!-- }); -->
<!--  -->
<!-- async function runWorker() { -->
<!--   await Worker.create({ -->
<!--     workflowsPath: require.resolve('./workflows'), -->
<!--   }); -->
<!-- } -->
<!--  -->
<!-- // Option 2: register in worker -->
<!-- // Pros: -->
<!-- // - Register activities here without worrying about conflicting workflow definitions -->
<!-- // - Works with existing SDK -->
<!--  -->
<!-- import * as workflows from './workflows'; -->
<!-- import * as activities from './activities'; -->
<!--  -->
<!-- async function runWorker() { -->
<!--   await Worker.create({ -->
<!--     workflowsPath: require.resolve('./workflows'), -->
<!--     operations: { -->
<!--       myOperation: { -->
<!--         workflow: workflows.http, -->
<!--         cancellable: true, -->
<!--       }, -->
<!--       mySignalOperation: { -->
<!--         workflow: workflows.unblockOrCancel, -->
<!--         signal: workflows.unblockSignal, -->
<!--       }, -->
<!--       myUpdateBasedOperation: { -->
<!--         workflow: workflows.unblockOrCancel, -->
<!--         update: workflows.someUpdateDefinition, -->
<!--       }, -->
<!--       myActivityOperation: { -->
<!--         activity: activities.someActivity, -->
<!--       }, -->
<!--     }, -->
<!--   }); -->
<!-- } -->
<!--  -->
<!-- // Option 3: workflow classes and decorators -->
<!-- // -->
<!-- // We may want to add this to the SDK regardless of Nexus -->
<!-- // -->
<!-- // Pros: -->
<!-- // - Operation definitions inline with workflow definitions -->
<!-- // Cons: -->
<!-- // - Operation definitions inline with workflow definitions -->
<!-- // - Doesn't apply to functional workflow style -->
<!-- // - Can be built in "user land" with option 2 (should this be another layer?) -->
<!--  -->
<!-- @Operation({ name: 'my-operation' }) -->
<!-- export class MyWorkflow { -->
<!--   constructor(input: MyInput) {} -->
<!--  -->
<!--   async execute() {} -->
<!--  -->
<!--   @Operation({ name: 'my-other-operation' }) -->
<!--   @Update() -->
<!--   async myUpdateHandler() {} -->
<!-- } -->
<!--  -->
