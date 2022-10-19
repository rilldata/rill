<script lang="ts">
  import { goto } from "$app/navigation";
  import { createEventDispatcher, getContext } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import type { Writable } from "svelte/store";
  import {
    ConnectorPropertyType,
    useRuntimeServiceMigrateSingle,
    V1Connector,
  } from "web-common/src/runtime-client";
  import type * as yup from "yup";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/layout-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
  import { getYupSchema } from "../../../connectors/schemas";
  import { Button } from "../../button";
  import AlertTriangle from "../../icons/AlertTriangle.svelte";
  import Input from "../../Input.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";

  export let connector: V1Connector;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const createSource = useRuntimeServiceMigrateSingle();

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const numTablesBeforeSubmit = $persistentTableStore.entities.length;

  const dispatch = createEventDispatcher();

  let yupSchema: yup.AnyObjectSchema;

  // state from svelte-forms-lib
  let form: Writable<any>;
  let errors: Writable<Record<never, string>>;
  let handleSubmit: (event: Event) => any;

  let waitingOnSourceImport = false;

  function compileCreateSourceSql(values) {
    const compiledKeyValues = Object.entries(values)
      .filter(([key]) => key !== "sourceName")
      .map(([key, value]) => `'${key}'='${value}'`)
      .join(", ");

    return (
      `CREATE SOURCE ${values.sourceName} WITH (connector = '${connector.name}', ` +
      compiledKeyValues +
      `)`
    );
  }

  function onConnectorChange(connector: V1Connector) {
    yupSchema = getYupSchema(connector);

    ({ form, errors, handleSubmit } = createForm({
      // TODO: initialValues should come from SQL asset and be reactive to asset modifications
      initialValues: {},
      // validationSchema: yupSchema, // removing temporarily, as it's preventing form submission
      onSubmit: (values) => {
        overlay.set({ title: `Importing ${values.sourceName}` });
        const sql = compileCreateSourceSql(values);
        // TODO: call runtime/repo.put() to create source artifact
        $createSource.mutate(
          {
            instanceId: runtimeInstanceId,
            data: { sql },
          },
          {
            onSuccess: async () => {
              waitingOnSourceImport = true;
              let numTables = numTablesBeforeSubmit;
              // poll the Node backend until it has picked up the new table in DuckDB
              while (numTables === numTablesBeforeSubmit) {
                numTables = $persistentTableStore.entities.length;
                await new Promise((resolve) => setTimeout(resolve, 1000));
              }

              const newSource = $persistentTableStore.entities.find(
                (entity) => entity.name === values.sourceName
              );

              waitingOnSourceImport = false;
              goto(`/source/${newSource.id}`);
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
</script>

{#if $createSource.isError}
  <div
    class="mx-4 my-2 p-2 flex bg-red-200 border-red-500 border-2 rounded text-red-800"
  >
    <AlertTriangle size="16px" />
    <p class="ml-2">
      {$createSource.error?.response?.data?.message}
    </p>
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
            error={$errors[property.key]}
            bind:value={$form[property.key]}
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
