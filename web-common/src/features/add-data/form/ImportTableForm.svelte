<script lang="ts">
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import TableSchema from "@rilldata/web-common/features/connectors/explorer/TableSchema.svelte";
  import {
    getAnalyzedConnectorByName,
    getAnalyzedConnectors,
  } from "@rilldata/web-common/features/connectors/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import { inferModelNameFromSQL } from "@rilldata/web-common/features/sources/sourceUtils.ts";
  import {
    type AddDataConfig,
    type ExploreConnectorStep,
    type ImportAddDataStepConfig,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    getConnectorDriverForSchema,
    getImportStepsForConnector,
  } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import ConnectorExplorer from "@rilldata/web-common/features/add-data/explorer/ConnectorExplorer.svelte";
  import type { ConnectorExplorerEntry } from "@rilldata/web-common/features/add-data/explorer/tree.ts";
  import { getLabelsForSource } from "@rilldata/web-common/features/add-data/form/form-labels.ts";
  import ResizableSidebar from "@rilldata/web-common/layout/ResizableSidebar.svelte";

  export let config: AddDataConfig;
  export let step: ExploreConnectorStep;
  export let onSubmit: (importConfig: ImportAddDataStepConfig) => void;

  const FormId = "import-table-form";

  const runtimeClient = useRuntimeClient();

  $: connectorDriverQuery = getAnalyzedConnectorByName(
    runtimeClient,
    step.connector,
  );
  $: connectorDriver =
    $connectorDriverQuery.data?.driver ??
    getConnectorDriverForSchema(step.schema);

  $: importSteps = connectorDriver
    ? getImportStepsForConnector(config, connectorDriver)
    : [];
  $: supportsModeling = importSteps[0] === ImportDataStep.CreateModel;

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

  const { form, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      resetForm: false,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        const sql =
          values.mode === "table"
            ? `SELECT * FROM ${values.schema ? values.schema + "." : ""}${values.table ?? ""}`
            : values.sql;
        const source =
          values.mode === "table"
            ? values.table
            : inferModelNameFromSQL(values.sql ?? "");
        if (!source) return; // TODO: error

        onSubmit({
          importSteps,
          source,
          sourceSchema: values.schema ?? "",
          sourceDatabase: values.database ?? "",
          connector: step.connector,
          sql,
          envBlob: null,
        } satisfies ImportAddDataStepConfig);
      },
      validationMethod: "onsubmit",
    },
  );
  $: isSubmitDisabled = (() => {
    if ($form.mode === "table") {
      return !$form.table;
    } else if ($form.mode === "sql") {
      return !$form.sql;
    }
    return false;
  })();

  $: connectors = getAnalyzedConnectors(runtimeClient, false);
  $: analyzedConnector = $connectors.data?.connectors?.find(
    (c) => c.name === step.connector,
  );

  $: sourceFormLabels = getLabelsForSource(importSteps);

  function handleTableChange({
    database,
    databaseSchema,
    table,
  }: ConnectorExplorerEntry) {
    if (!table) return;
    form.update((f) => {
      f.database = database;
      f.schema = databaseSchema;
      f.table = table;
      return f;
    });
  }
</script>

<form
  use:enhance
  on:submit|preventDefault={submit}
  id={FormId}
  class="flex flex-col gap-1 h-full overflow-y-auto"
>
  <div class="flex flex-col gap-2 px-6 pt-2">
    {#if supportsModeling}
      <div>
        Pick a table or input your file SQL to power your first dashboard
      </div>
      <Tabs bind:value={$form["mode"]} options={modeOptions} disableMarginTop>
        {#each modeOptions as option (option.value)}
          <TabsContent value={option.value} />
        {/each}
      </Tabs>
    {:else}
      <div>Pick a table to power your first dashboard</div>
    {/if}
  </div>
  {#if $form["mode"] === "table"}
    {#if analyzedConnector}
      <div class="flex flex-row size-full overflow-hidden border-t">
        <div class="flex-grow border-r ml-6 mt-2">
          <ConnectorExplorer
            connectorName={step.connector}
            onSelect={handleTableChange}
          />
        </div>
        <ResizableSidebar
          id="table-schema-sidebar"
          minWidth={100}
          maxWidth={500}
          defaultWidth={288}
          additionalClass="overflow-auto bg-surface-subtle p-2"
        >
          {#if $form["table"]}
            <TableSchema
              connector={step.connector}
              database={$form["database"]}
              databaseSchema={$form["schema"]}
              table={$form["table"]}
              addLeftPadding={false}
            />
          {/if}
        </ResizableSidebar>
      </div>
    {/if}
  {:else if $form["mode"] === "sql"}
    <div class="flex-grow px-6">
      <Input id="sql" label="SQL" bind:value={$form["sql"]} />
    </div>
  {/if}

  <div class="flex flex-row px-6 py-4 gap-2 border-t">
    <Button onClick={() => window.history.back()} type="secondary">Back</Button>
    <div class="grow" />
    <Button
      disabled={$submitting || isSubmitDisabled}
      loading={$submitting}
      onClick={submit}
      type="primary"
    >
      {sourceFormLabels.primaryButtonLabel}
    </Button>
  </div>
</form>
