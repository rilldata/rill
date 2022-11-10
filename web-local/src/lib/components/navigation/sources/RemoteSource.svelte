<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    ConnectorProperty,
    ConnectorPropertyType,
    getRuntimeServiceListCatalogObjectsQueryKey,
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
  import { Button } from "../../button";
  import InformationalField from "../../forms/InformationalField.svelte";
  import Input from "../../forms/Input.svelte";
  import SubmissionError from "../../forms/SubmissionError.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";
  import { humanReadableErrorMessage } from "./errors";
  import {
    compileCreateSourceSql,
    inferSourceName,
    waitForSource,
  } from "./sourceUtils";
  import {
    fromYupFriendlyKey,
    getYupSchema,
    toYupFriendlyKey,
  } from "./yupSchemas";

  export let connector: V1Connector;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  const dispatch = createEventDispatcher();

  let connectorProperties: ConnectorProperty[];
  let yupSchema: yup.AnyObjectSchema;

  // state from svelte-forms-lib
  let form: Writable<any>;
  let touched: Writable<Record<any, boolean>>;
  let errors: Writable<Record<never, string>>;
  let handleChange: (event: Event) => any;
  let handleSubmit: (event: Event) => any;

  let waitingOnSourceImport = false;

  function onConnectorChange(connector: V1Connector) {
    yupSchema = getYupSchema(connector);

    ({ form, touched, errors, handleChange, handleSubmit } = createForm({
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
                getRuntimeServiceListCatalogObjectsQueryKey(runtimeInstanceId, {
                  type: RuntimeServiceListCatalogObjectsType.TYPE_SOURCE,
                })
              );
            },
            onError: () => {
              overlay.set(null);
            },
          }
        );
      },
    }));

    // Place the "Source name" field directly under the "Path" field, which is the first property for each connector (s3, gcs, https).
    connectorProperties = [
      ...connector.properties.slice(0, 1),
      {
        key: "sourceName",
        displayName: "Source name",
        description: "The name of the source",
        placeholder: "my_new_source",
        type: ConnectorPropertyType.TYPE_STRING,
        nullable: false,
      },
      ...connector.properties.slice(1),
    ];
  }

  $: onConnectorChange(connector);

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($touched.sourceName) return;
      const sourceName = inferSourceName(connector, value);
      $form.sourceName = sourceName ? sourceName : $form.sourceName;
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
        message={humanReadableErrorMessage(
          connector.name,
          $createSource.error.response.data.code,
          $createSource.error.response.data.message
        )}
      />
    {/if}

    {#each connectorProperties as property}
      {@const label =
        property.displayName + (property.nullable ? " (optional)" : "")}
      <div class="py-2">
        {#if property.type === ConnectorPropertyType.TYPE_STRING}
          <Input
            id={toYupFriendlyKey(property.key)}
            {label}
            placeholder={property.placeholder}
            hint={property.hint}
            error={$errors[toYupFriendlyKey(property.key)]}
            bind:value={$form[toYupFriendlyKey(property.key)]}
            on:input={onStringInputChange}
            on:change={handleChange}
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
