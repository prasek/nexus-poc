## Hand coded Nexus handler

### Lowest handler level

#### Sync request - "raw handler"

`worker.ts`

```ts
import { Worker, NexusWorker } from "@temporalio/worker";
import { ApplicationFailure } from "@temporalio/common";
import { Request, Response } from "@temporalio/nexus"; // name TBD

await Worker.create({
  namespace: "foo",
  taskQueue: "my-tq-for-workflows-and-activities",
  workflowsPath: require.resolve("./workflows"),
  activities,
});

await NexusWorker.create({
  namespace: "foo",
  service: "calculator",
  async start({ operation, headers, body }: Request): Promise<Response> {
    // handler is essentially a router.
    if (operation !== "add") {
      // Nexus has a notion of retryable and non-retryable errors.
      // Non-`ApplicationFailure` errors thrown in a handler are converted to retryable `ApplicationFailure`s.
      // TODO: Decide if we want to reuse the ApplicationFailure class here.
      throw ApplicationFailure.nonRetryable("Not found");
    }
    if (!headers["content-type"].startsWith("application/json")) {
      throw ApplicationFailure.nonRetryable("Invalid content type");
    }

    const { x, y } = JSON.parse(body);
    const result = x + y;

    return {
      headers: { "Content-Type": "application/json" },
      result: {
        completed: {
          succeeded: {
            // TODO: Decide whether the headers should be close to the body or part of the response object.
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ result }),
          },
        },
      },
    };
  },
});
```

#### Start operation

```ts
await NexusWorker.create({
  namespace: "foo",
  service: "payments",
  async handler({
    method,
    operation,
    id,
    headers,
    body,
    callbackURL,
  }: Request): Promise<Response> {
    if (headers["authorization"] !== "Bearer top-secret") {
      // Later we may add an HTTP gateway in front of Temporal (At a later stage in the project).
      // When we get there, we'll need to support setting the status code (as well as other HTTP specific features) in handlers.
      // HTTP specific data may be delivered in specific response headers.
      throw new ApplicationFailure.nonRetryable("Unauthorized");
    }
    if (operation === "checkout") {
      if (!headers["content-type"].startsWith("application/json")) {
        throw ApplicationFailure.nonRetryable("Invalid content type");
      }
      const workflowId = await qualifyOperationId(operation, id, headers);

      const { product } = JSON.parse(body);
      const handle = await client.workflow.start(workflows.payment, {
        workflowId,
        args: [{ customerId, amount }],
        callbackURL,
      });
      return {
        outcome: {
          started: {
            callbackURLSupported: true,
          },
        },
      };
    }
  },
});

async function qualifyOperationId(operation, id, headers) {
  const customerId = await getCustomerInfo(headers);
  const qualifiedId = `${operation}-${customerId}-${id}`;
  // Option 1 - use client Id directly, qualified for multi-tenancy
  return qualifiedId;
  // Option 2 - use a persistent mapping (e.g. when operation is backed by an update and requires two identifiers to
  address).
  return await getClientToHandlerIdMappingFromTemporal(qualifiedId);
}
```

#### Data conversion

// TODO: complete me

```ts
async function handler({
  method,
  operation,
  headers,
  body,
}: Request): Promise<Response> {
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
      headers: { "Content-Type": "application/json" },
      outcome: {
        started: {
          callbackURLSupported: true,
        },
      },
    };
  }
}
```

### Cancel an operation

### Bindings to workflows

### Middleware

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
