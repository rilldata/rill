<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    compileCreateSourceSql,
    waitForSource,
  } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
  import { createEventDispatcher, getContext } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import type { Writable } from "svelte/store";
  import {
    ConnectorPropertyType,
    RpcStatus,
    useRuntimeServiceMigrateSingle,
    V1Connector,
  } from "web-common/src/runtime-client";
  import type * as yup from "yup";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/layout-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
  import {
    fromYupFriendlyKey,
    getYupSchema,
    toYupFriendlyKey,
  } from "../../../connectors/schemas";
  import { Button } from "../../button";
  import Input from "../../forms/Input.svelte";
  import SubmissionError from "../../forms/SubmissionError.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";

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
            },
            onError: (error) => {
              console.error(error);
              overlay.set(null);
            },
          }
        );
      },
    }));
  }

  $: onConnectorChange(connector);

  function humanReadableErrorMessage(error: RpcStatus) {
    // TODO: the error response type does not match the type defined in the API
    switch (error.response.data.code) {
      // gRPC error codes: https://pkg.go.dev/google.golang.org/grpc@v1.49.0/codes
      case 3:
        // InvalidArgument
        return error.response.data.message;
      default:
        return "An unknown error occurred. If the error persists, please reach out for help on <a href=https://bit.ly/3unvA05 target=_blank>Discord</a>.";
    }
  }
</script>

{#if $createSource.isError}
  <div class="mx-4">
    <SubmissionError message={humanReadableErrorMessage($createSource.error)} />
  </div>
{/if}

<div class="px-4 flex-grow overflow-y-auto pb-2">
  <form
    on:submit|preventDefault={handleSubmit}
    id="remote-source-{connector}-form"
  >
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
            label={property.displayName}
            placeholder={property.placeholder}
            hint={property.hint}
            error={$errors[toYupFriendlyKey(property.key)]}
            bind:value={$form[toYupFriendlyKey(property.key)]}
          />
        {/if}
        {#if property.type === ConnectorPropertyType.TYPE_BOOLEAN}
          <label for={property.key} class="flex items-center">
            <input
              id={property.key}
              type="checkbox"
              bind:checked={$form[property.key]}
              class="h-5 w-5"
            />
            <span class="ml-2 text-sm">{label}</span>
          </label>
        {/if}
      </div>
    {/each}
  </form>
</div>
<div class="bg-gray-100 border-t border-gray-300">
  <DialogFooter>
    <div class="flex items-center space-x-2">
      <Button
        type="primary"
        submitForm
        form="remote-source-{connector}-form"
        disabled={$createSource.isLoading || waitingOnSourceImport}
      >
        Add source
      </Button>
    </div>
  </DialogFooter>
</div>
