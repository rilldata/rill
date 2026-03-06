<script lang="ts">
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import DatabaseExplorer from "@rilldata/web-common/features/connectors/explorer/DatabaseExplorer.svelte";
  import TableSchema from "@rilldata/web-common/features/connectors/explorer/TableSchema.svelte";
  import { ConnectorExplorerStore } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store.ts";
  import { getAnalyzedConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import {
    compileSourceYAML,
    inferModelNameFromSQL,
  } from "@rilldata/web-common/features/sources/sourceUtils.ts";
  import {
    type AddDataConfig,
    type ImportAddDataStepConfig,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getImportStepsForConnector } from "@rilldata/web-common/features/add-data/steps/transitions.ts";

  export let config: AddDataConfig;
  export let connectorName: string;
  export let connectorDriver: V1ConnectorDriver;
  export let onSubmit: (importConfig: ImportAddDataStepConfig) => void;

  const FormId = "import-table-form";

  const runtimeClient = useRuntimeClient();

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
    mode: string;
    table: string;
    database: string;
    schema: string;
    sql: string;
  } = {
    mode: modeOptions[0].value,
    table: "",
    database: "",
    schema: "",
    sql: "",
  };
  const schema = yup(
    object({
      mode: string().required(),
      table: string().when("mode", {
        is: "table",
        then: (schema) => schema.required(),
        otherwise: (schema) => schema.notRequired(),
      }),
      database: string(),
      schema: string(),
      sql: string().when("mode", {
        is: "sql",
        then: (schema) => schema.required(),
        otherwise: (schema) => schema.notRequired(),
      }),
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
        values.mode === "table"
          ? `SELECT * FROM ${values.table ?? ""}`
          : values.sql;
      const name =
        values.mode === "table"
          ? values.table
          : inferModelNameFromSQL(values.sql ?? "");
      if (!name) return; // TODO: error

      const modelName = getName(
        name,
        fileArtifacts.getNamesForKind(ResourceKind.Model),
      );
      const yaml = compileSourceYAML(
        connectorDriver,
        {
          name: modelName,
          sql,
          database: values.database,
        },
        {
          connectorInstanceName: connectorName,
        },
      );

      onSubmit({
        importSteps: getImportStepsForConnector(config, connectorDriver),
        source: modelName,
        sourceSchema: values.schema ?? "",
        sourceDatabase: values.database ?? "",
        connector: connectorName,
        yaml,
        envBlob: null,
      } satisfies ImportAddDataStepConfig);
    },
    validationMethod: "onsubmit",
  });

  $: connectors = getAnalyzedConnectors(runtimeClient, false);
  $: analyzedConnector = $connectors.data?.connectors?.find(
    (c) => c.name === connectorName,
  );

  $: connectorExplorerStore = new ConnectorExplorerStore(
    {
      allowContextMenu: false,
      allowNavigateToTable: false,
      allowSelectTable: false,
      allowShowSchema: false,
    },
    (_, database, schema, table) => {
      if (!database || !schema || !table) return;
      form.update((f) => {
        f.database = database;
        f.schema = schema;
        f.table = table;
        return f;
      });
    },
  );
</script>

<form
  use:enhance
  on:submit|preventDefault={submit}
  id={FormId}
  class="flex flex-col gap-1 h-full"
>
  <div class="flex flex-col gap-2 px-6 pt-2">
    <div>Pick a table or Input your file SQL to power your first dashboard</div>
    <Tabs bind:value={$form["mode"]} options={modeOptions} disableMarginTop>
      {#each modeOptions as option (option.value)}
        <TabsContent value={option.value} />
      {/each}
    </Tabs>
  </div>
  <div class="grow">
    {#if $form["mode"] === "table"}
      {#if analyzedConnector}
        <div class="flex flex-row size-full border-t">
          <div class="grow border-r">
            <DatabaseExplorer
              connector={analyzedConnector}
              store={connectorExplorerStore}
            />
          </div>
          <div class="bg-surface-subtle w-[40%] p-2">
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
      <div class="px-6">
        <Input id="sql" label="SQL" bind:value={$form["sql"]} />
      </div>
    {/if}
  </div>

  <div class="flex flex-row px-6 py-4 gap-2 border-t">
    <Button onClick={() => window.history.back()} type="secondary">Back</Button>
    <div class="grow" />
    <Button onClick={submit} type="primary">Generate dashboard with AI</Button>
  </div>
</form>
