<script lang="ts">
  import { page } from "$app/stores";
  import NewSourceSelector from "@rilldata/web-common/features/add-data/form/NewSourceSelector.svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import SourceForm from "@rilldata/web-common/features/add-data/form/SourceForm.svelte";
  import ImportTableForm from "@rilldata/web-common/features/add-data/form/ImportTableForm.svelte";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
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

  export let config: AddDataConfig;

  let stepState: AddDataState = { step: AddDataStep.Select };

  // beforeNavigate, onNavigate or afterNavigate do not seem to get called when state changes.
  // "popstate" event does not have direct access to the page state, rather it is under `sveltekit:states` key which seems like internal key.
  // So we need this reactive statement to update the state.
  $: if (($page.state as AddDataState).step) {
    stepState = $page.state as AddDataState;
  }

  $: schema = (stepState as any).schema as string | undefined;
  $: connector = (stepState as any).connector as string | undefined;
  $: connectorDriver = maybeGetConnectorDriver(
    config.instanceId,
    schema,
    connector,
  );

  $: displayIcon =
    connectorIconMapping[connector ?? ""] ??
    connectorIconMapping[connectorDriver?.name ?? ""];
  $: displayName = connectorDriver?.displayName ?? connector;

  function transitionToSchema(schema: string) {
    const newState = transitionToNextStep(config, stepState, {
      schema: schema,
    });
    console.log("transition:schema", newState);
    pushState("", newState);
  }

  function transitionToConnector(connector: string) {
    const newState = transitionToNextStep(config, stepState, {
      connector: connector,
    });
    console.log("transition:connector", newState);
    pushState("", newState);
  }

  function setAndStartImport(importConfig: ImportAddDataStepConfig) {
    const newState = transitionToNextStep(config, stepState, {
      importConfig,
    });
    console.log("transition:source/explorer", newState);
    pushState("", newState);
  }
</script>

<div
  class="flex flex-col size-full bg-surface-background border rounded-lg shadow-sm;"
>
  {#if displayName && stepState.step !== AddDataStep.Import}
    <div class="flex flex-row items-center px-6 py-4 gap-1 border-b">
      {#if displayIcon}
        <svelte:component this={displayIcon} size="18px" />
      {/if}
      <span class="text-lg leading-none font-semibold">{displayName}</span>
    </div>
  {/if}
  {#if stepState.step === AddDataStep.Select}
    <NewSourceSelector onSelect={transitionToSchema} />
  {:else if stepState.step === AddDataStep.Connector}
    {#if connectorDriver}
      <ConnectorForm
        {connectorDriver}
        onSubmit={transitionToConnector}
        onBack={() => window.history.back()}
      />
    {:else}
      <div>No connector driver (TODO)</div>
    {/if}
  {:else if stepState.step === AddDataStep.Source}
    {#if connectorDriver && schema && connector}
      <SourceForm
        {connectorDriver}
        schemaName={schema}
        connectorName={connector}
        onSubmit={setAndStartImport}
        onBack={() => window.history.back()}
      />
    {:else}
      <div>Missing connector driver, schema name, or connector name (TODO)</div>
    {/if}
  {:else if stepState.step === AddDataStep.Explorer}
    {#if connectorDriver && connector}
      <ImportTableForm
        {connectorDriver}
        connectorName={connector}
        onSubmit={setAndStartImport}
      />
    {:else}
      <div>Missing connector driver or connector name (TODO)</div>
    {/if}
  {:else if stepState.step === AddDataStep.Import}
    <ImportTableStatus
      {config}
      importAddDataStep={stepState}
      onBack={() => window.history.back()}
    />
  {/if}
</div>
