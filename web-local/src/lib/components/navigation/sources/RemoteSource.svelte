<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    ConnectorPropertyType,
    getRuntimeServiceListCatalogObjectsQueryKey,
    RpcStatus,
    RuntimeServiceListCatalogObjectsType,
    useRuntimeServiceMigrateSingle,
    V1Connector,
  } from "@rilldata/web-common/runtime-client";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { createEventDispatcher, getContext } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import type { Writable } from "svelte/store";
  import type * as yup from "yup";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
  import {
    fromYupFriendlyKey,
    getYupSchema,
    toYupFriendlyKey,
  } from "../../../connectors/schemas";
  import { Button } from "../../button";
  import InformationalField from "../../forms/InformationalField.svelte";
  import Input from "../../forms/Input.svelte";
  import SubmissionError from "../../forms/SubmissionError.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";
  import { compileCreateSourceSql, waitForSource } from "./sourceUtils";

  export let connector: V1Connector;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  const dispatch = createEventDispatcher();

  let yupSchema: yup.AnyObjectSchema;

  // state from svelte-forms-lib
  let form: Writable<any>;
  let errors: Writable<Record<never, string>>;
  let handleSubmit: (event: Event) => any;

  let waitingOnSourceImport = false;

  function onConnectorChange(connector: V1Connector) {
    yupSchema = getYupSchema(connector);

    ({ form, errors, handleSubmit } = createForm({
      // TODO: initialValues should come from SQL asset and be reactive to asset modifications
      initialValues: {
        sourceName: "", // avoids `values.sourceName` warning
      },
      validationSchema: yupSchema,
      onSubmit: (values) => {
        overlay.set({ title: `Importing ${values.sourceName}` });
        const formValues = Object.fromEntries(
          Object.entries(values).map(([key, value]) => [
            fromYupFriendlyKey(key),
            value,
          ])
        );

        const sql = compileCreateSourceSql(formValues, connector.name);
        // TODO: call runtime/repo.put() to create source artifact
        $createSource.mutate(
          {
            instanceId: runtimeInstanceId,
            data: { sql, createOrReplace: false },
          },
          {
            onSuccess: async () => {
              waitingOnSourceImport = true;
              const newId = await waitForSource(
                values.sourceName,
                persistentTableStore
              );
              waitingOnSourceImport = false;
              goto(`/source/${newId}`);
              dispatch("close");
              overlay.set(null);
              return queryClient.invalidateQueries(
                getRuntimeServiceListCatalogObjectsQueryKey(runtimeInstanceId, { type: RuntimeServiceListCatalogObjectsType.TYPE_SOURCE })
              );
            },
            onError: () => {
              overlay.set(null);
            },
          }
        );
      },
    }));
  }

  $: onConnectorChange(connector);

  function humanReadableErrorMessage(connectorName: string, error: RpcStatus) {
    // TODO: the error response type does not match the type defined in the API
    switch (error.response.data.code) {
      // gRPC error codes: https://pkg.go.dev/google.golang.org/grpc@v1.49.0/codes
      // InvalidArgument
      case 3: {
        const serverError = error.response.data.message;

        // Rill errors
        if (
          serverError.match(/an existing object with name '.*' already exists/)
        ) {
          return "A source with this name already exists. Please choose a different name.";
        }

        // AWS errors (ref: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html)
        if (connectorName === "s3") {
          if (serverError.includes("MissingRegion")) {
            return "Region not detected. Please enter a region.";
          } else if (serverError.includes("NoCredentialProviders")) {
            return "No credentials found. Please see the docs for how to configure AWS credentials.";
          } else if (serverError.includes("InvalidAccessKey")) {
            return "Invalid AWS access key. Please check your credentials.";
          } else if (serverError.includes("SignatureDoesNotMatch")) {
            return "Invalid AWS secret key. Please check your credentials.";
          } else if (serverError.includes("BucketRegionError")) {
            return "Bucket is not in the provided region. Please check your region.";
          } else if (serverError.includes("AccessDenied")) {
            return "Access denied. Please ensure you have the correct permissions.";
          } else if (serverError.includes("NoSuchKey")) {
            return "Invalid path. Please check your path.";
          } else if (serverError.includes("NoSuchBucket")) {
            return "Invalid bucket. Please check your bucket name.";
          } else if (serverError.includes("AuthorizationHeaderMalformed")) {
            return "Invalid authorization header. Please check your credentials.";
          }
        }

        // GCP errors (ref: https://cloud.google.com/storage/docs/json_api/v1/status-codes)
        if (connectorName === "gcs") {
          if (serverError.includes("could not find default credentials")) {
            return "No credentials found. Please see the docs for how to configure GCP credentials.";
          } else if (serverError.includes("Unauthorized")) {
            return "Unauthorized. Please check your credentials.";
          } else if (serverError.includes("AccessDenied")) {
            return "Access denied. Please ensure you have the correct permissions.";
          }
        }

        if (connectorName === "https") {
          if (serverError.includes("invalid file")) {
            return "The provided URL does not appear to have a valid dataset. Please check your path and try again.";
          } else if (serverError.includes("failed to fetch url")) {
            return "We could not connect to the provided URL. Please check your path and try again.";
          }
        }

        // DuckDB errors
        if (serverError.match(/expected \d* values per row, but got \d*/)) {
          return "Malformed CSV file: number of columns does not match header.";
        } else if (
          serverError.match(/Catalog Error: Table with name .* does not exist/)
        ) {
          return "We had trouble ingesting your data. Please see the docs for common issues. If you're still stuck, don't hesitate to reach out on Discord.";
        }
        return error.response.data.message;
      }
      default:
        return "An unknown error occurred. If the error persists, please reach out for help on <a href=https://bit.ly/3unvA05 target=_blank>Discord</a>.";
    }
  }
</script>

<div class="h-full flex flex-col">
  <form
    on:submit|preventDefault={handleSubmit}
    id="remote-source-{connector.name}-form"
    class="px-4 pb-2 flex-grow overflow-y-auto"
  >
    <div class="pt-4 pb-2">
      Need help? Refer to our
      <a href="https://docs.rilldata.com/import-data" target="_blank">docs</a> for
      more information.
    </div>
    {#if $createSource.isError}
      <SubmissionError
        message={humanReadableErrorMessage(connector.name, $createSource.error)}
      />
    {/if}
    <div class="py-2">
      <Input
        label="Source name"
        bind:value={$form["sourceName"]}
        error={$errors["sourceName"]}
        placeholder="my_new_source"
      />
    </div>
    {#each connector.properties as property}
      {@const label =
        property.displayName + (property.nullable ? " (optional)" : "")}
      <div class="py-2">
        {#if property.type === ConnectorPropertyType.TYPE_STRING}
          <Input
            id={property.key}
            {label}
            placeholder={property.placeholder}
            hint={property.hint}
            error={$errors[toYupFriendlyKey(property.key)]}
            bind:value={$form[toYupFriendlyKey(property.key)]}
          />
        {:else if property.type === ConnectorPropertyType.TYPE_BOOLEAN}
          <label for={property.key} class="flex items-center">
            <input
              id={property.key}
              type="checkbox"
              bind:checked={$form[property.key]}
              class="h-5 w-5"
            />
            <span class="ml-2 text-sm">{label}</span>
          </label>
        {:else if property.type === ConnectorPropertyType.TYPE_INFORMATIONAL}
          <InformationalField
            description={property.description}
            hint={property.hint}
            href={property.href}
          />
        {/if}
      </div>
    {/each}
  </form>
  <div class="bg-gray-100 border-t border-gray-300">
    <DialogFooter>
      <div class="flex items-center space-x-2">
        <Button
          type="primary"
          submitForm
          form="remote-source-{connector.name}-form"
          disabled={$createSource.isLoading || waitingOnSourceImport}
        >
          Add source
        </Button>
      </div>
    </DialogFooter>
  </div>
</div>
