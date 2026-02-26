<script lang="ts">
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import DatabaseExplorer from "@rilldata/web-common/features/connectors/explorer/DatabaseExplorer.svelte";
  import TableSchema from "@rilldata/web-common/features/connectors/explorer/TableSchema.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import {
    ConnectorExplorerStore,
    type ConnectorTableEntry,
  } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store.ts";
  import { getAnalyzedConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { compileSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils.ts";
  import { get } from "svelte/store";

  export let connectorName: string;
  export let connectorDriver: V1ConnectorDriver;
  export let onCreate: (
    name: string,
    tableEntry: ConnectorTableEntry,
    yaml: string,
  ) => void;

  const FormId = "import-table-form";

  const modeOptions = [
    {
      label: "Table",
      value: "table",
    },
    {
      label: "SQL",
      value: "sql",
    },
  ];

  const initialValues: {
    name: string;
    mode: string;
    table: string;
    database: string;
    schema: string;
    sql: string;
  } = {
    name: "",
    mode: modeOptions[0].value,
    table: "",
    database: "",
    schema: "",
    sql: "",
  };
  const schema = yup(
    object({
      name: string().required(),
      mode: string().required(),
      table: string(),
      database: string(),
      schema: string(),
      sql: string(),
    }),
  );

  const { form, enhance, submit } = superForm(defaults(initialValues, schema), {
    SPA: true,
    validators: schema,
    resetForm: false,
    async onUpdate({ form }) {
      if (!form.valid) return;
      const values = form.data;
      const sql =
        values.mode === "table" ? `SELECT * FROM ${values.table}` : values.sql;
      const yaml = compileSourceYAML(connectorDriver, {
        name: values.name,
        sql,
        database: values.database,
      });
      onCreate(
        values.name,
        get(connectorExplorerStore.selectedTableStore),
        yaml,
      );
    },
    validationMethod: "onsubmit",
  });

  $: ({ instanceId } = $runtime);
  $: connectors = getAnalyzedConnectors(instanceId, false);
  $: analyzedConnector = $connectors.data?.connectors?.find(
    (c) => c.name === connectorName,
  );

  let nameChangedDirectly = false;

  $: connectorExplorerStore = new ConnectorExplorerStore(
    {
      allowContextMenu: false,
      allowNavigateToTable: false,
      allowSelectTable: false,
      allowShowSchema: false,
    },
    (_, database, schema, table) => {
      if (!database || !schema || !table) return;
      connectorExplorerStore.selectedTableStore.set({
        connector: connectorName,
        database,
        schema,
        table,
      });

      form.update((f) => {
        f.database = database;
        f.schema = schema;
        f.table = table;
        if (!nameChangedDirectly) {
          f.name = getName(
            table,
            fileArtifacts.getNamesForKind(ResourceKind.Model),
          );
        }
        return f;
      });
    },
  );
</script>

<form use:enhance on:submit|preventDefault={submit} id={FormId}>
  <Input
    id="name"
    label="Name"
    bind:value={$form["name"]}
    onInput={() => (nameChangedDirectly = true)}
  />
  <div>Pick a table or Input your file SQL to power your first dashboard</div>
  <Tabs bind:value={$form["mode"]} options={modeOptions} disableMarginTop>
    {#each modeOptions as option (option.value)}
      <TabsContent value={option.value} />
    {/each}
  </Tabs>
  {#if $form["mode"] === "table"}
    {#if analyzedConnector}
      <div class="flex flex-row gap-2 w-full">
        <DatabaseExplorer
          {instanceId}
          connector={analyzedConnector}
          store={connectorExplorerStore}
        />
        <div class="bg-surface-subtle">
          {#if $form["table"]}
            <TableSchema
              connector={connectorName}
              database={$form["database"]}
              databaseSchema={$form["schema"]}
              table={$form["table"]}
              addLeftPadding={false}
            />
          {/if}
        </div>
      </div>
    {/if}
  {:else if $form["mode"] === "sql"}
    <Input id="sql" label="SQL" bind:value={$form["sql"]} />
  {/if}

  <div class="flex flex-row mt-4 gap-2">
    <Button onClick={() => window.history.back()} type="secondary">Back</Button>
    <div class="grow" />
    <Button onClick={submit} type="primary">Generate dashboard with AI</Button>
  </div>
</form>
