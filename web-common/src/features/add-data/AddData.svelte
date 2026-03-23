<script lang="ts">
  import { page } from "$app/stores";
  import SourceSelector from "@rilldata/web-common/features/add-data/form/SourceSelector.svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import SourceForm from "@rilldata/web-common/features/add-data/form/SourceForm.svelte";
  import ImportTableForm from "@rilldata/web-common/features/add-data/form/ImportTableForm.svelte";
  import ImportTableStatus from "@rilldata/web-common/features/add-data/ImportTableStatus.svelte";
  import {
    type AddDataConfig,
    AddDataStep,
    type AddDataState,
    type ImportAddDataStepConfig,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import {
    maybeGetConnectorDriver,
    transitionToNextStep,
  } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import { pushState } from "$app/navigation";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import ConnectorHeader from "@rilldata/web-common/features/add-data/form/ConnectorHeader.svelte";

  export let config: AddDataConfig = {};
  export let onClose: () => void = () => {};

  const runtimeClient = useRuntimeClient();

  let stepState: AddDataState = { step: AddDataStep.SelectConnector };

  // beforeNavigate, onNavigate or afterNavigate do not seem to get called when state changes.
  // "popstate" event does not have direct access to the page state, rather it is under `sveltekit:states` key which seems like internal key.
  // So we need this reactive statement to update the state.
  $: if ("step" in $page.state) {
    stepState = $page.state as AddDataState;
  }
  $: console.log(stepState);

  $: schema = (stepState as any).schema as string | undefined;
  $: connector = (stepState as any).connector as string | undefined;
  let connectorDriver: V1ConnectorDriver | null = null;
  $: void maybeGetConnectorDriver(runtimeClient, schema, connector).then(
    (d) => (connectorDriver = d),
  );

  $: isImportStep = stepState.step === AddDataStep.Import;
  $: sizeClass = isImportStep ? "h-fit w-[500px]" : "h-[630px] w-[900px]";
  $: shouldShowHeader =
    stepState.step === AddDataStep.CreateConnector ||
    stepState.step === AddDataStep.CreateModel ||
    stepState.step === AddDataStep.ExploreConnector;

  async function transitionFromInit(
    schema: string | undefined,
    connector: string | undefined,
  ) {
    const newState = await transitionToNextStep(
      runtimeClient,
      {
        step: AddDataStep.SelectConnector,
      },
      {
        schema,
        connector,
      },
    );
    pushState("", newState);
  }

  async function transitionToSchema(schema: string) {
    const newState = await transitionToNextStep(runtimeClient, stepState, {
      schema,
    });
    pushState("", newState);
  }

  async function setAndStartImport(importConfig: ImportAddDataStepConfig) {
    const newState = await transitionToNextStep(runtimeClient, stepState, {
      importConfig,
    });
    pushState("", newState);
  }
</script>

<div
  class="flex flex-col {sizeClass} bg-surface-background border rounded-lg shadow-sm"
>
  {#if shouldShowHeader && schema}
    <ConnectorHeader
      {config}
      schemaName={schema}
      connectorName={connector}
      onConnectorChange={(newConnector) =>
        transitionFromInit(undefined, newConnector)}
      onNewConnector={() => transitionFromInit(schema, undefined)}
    />
  {/if}

  {#if stepState.step === AddDataStep.SelectConnector}
    <SourceSelector
      {config}
      onSelect={transitionToSchema}
      onBack={() => window.history.back()}
    />
  {:else if stepState.step === AddDataStep.CreateConnector}
    <ConnectorForm
      step={stepState}
      onSubmit={(newState) => pushState("", newState)}
      onBack={() => window.history.back()}
      {onClose}
    />
  {:else if stepState.step === AddDataStep.CreateModel}
    {#key stepState.connector}
      <SourceForm
        {config}
        step={stepState}
        onSubmit={setAndStartImport}
        onBack={() => window.history.back()}
      />
    {/key}
  {:else if stepState.step === AddDataStep.ExploreConnector}
    {#key stepState.connector}
      <ImportTableForm {config} step={stepState} onSubmit={setAndStartImport} />
    {/key}
  {:else if stepState.step === AddDataStep.Import}
    <ImportTableStatus
      importAddDataStep={stepState}
      onBack={() => window.history.back()}
      {onClose}
    />
  {/if}
</div>
