<script lang="ts">
  import { page } from "$app/stores";
  import {
    AddDataManager,
    AddDataStep,
  } from "@rilldata/web-common/features/add-data/AddDataManager.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { get } from "svelte/store";
  import NewSourceSelector from "@rilldata/web-common/features/add-data/form/NewSourceSelector.svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import SourceForm from "@rilldata/web-common/features/add-data/form/SourceForm.svelte";
  import ImportTableForm from "@rilldata/web-common/features/add-data/import/ImportTableForm.svelte";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  export let initSchemaName: string | null = null;
  export let initConnectorName: string | null = null;

  $: manager = new AddDataManager(
    get(runtime).instanceId,
    initSchemaName,
    initConnectorName,
  );
  $: ({ stepStore, connectorDriverStore, schemaNameStore, connectorNameStore } =
    manager);

  // beforeNavigate, onNavigate or afterNavigate do not seem to get called when state changes.
  // "popstate" event does not have direct access to the page state, rather it is under `sveltekit:states` key which seems like internal key.
  // So we need this reactive statement to update the state.
  $: manager.applyState($page.state);

  $: step = $stepStore;
  $: connectorDriver = $connectorDriverStore;
  $: schemaName = $schemaNameStore;
  $: connectorName = $connectorNameStore;

  $: displayIcon =
    connectorIconMapping[connectorName] ??
    connectorIconMapping[connectorDriver?.name ?? ""];
  $: displayName = connectorDriver?.displayName ?? connectorName;
</script>

<div
  class="flex flex-col gap-y-4 w-full bg-surface-background border rounded-lg shadow-sm;"
>
  {#if displayName}
    <div class="flex flex-row items-center px-6 py-4 gap-1 border-b">
      {#if displayIcon}
        <svelte:component this={displayIcon} size="18px" />
      {/if}
      <span class="text-lg leading-none font-semibold">{displayName}</span>
    </div>
  {/if}
  {#if step === AddDataStep.Select}
    <NewSourceSelector
      onSelect={(newSchemaName) => manager.selectSchemaName(newSchemaName)}
    />
  {:else if step === AddDataStep.Connector}
    {#if connectorDriver}
      <ConnectorForm
        {connectorDriver}
        onSubmit={(newConnectorName) =>
          manager.selectConnector(schemaName, newConnectorName)}
        onBack={() => window.history.back()}
      />
    {:else}
      <div>No connector driver (TODO)</div>
    {/if}
  {:else if step === AddDataStep.Source}
    {#if connectorDriver && schemaName && connectorName}
      <SourceForm
        {connectorDriver}
        {schemaName}
        {connectorName}
        onSubmit={() => {}}
        onBack={() => window.history.back()}
      />
    {:else}
      <div>Missing connector driver, schema name, or connector name (TODO)</div>
    {/if}
  {:else if step === AddDataStep.Explorer}
    {#if connectorDriver && connectorName}
      <ImportTableForm {connectorDriver} {connectorName} onCreate={() => {}} />
    {:else}
      <div>Missing connector driver or connector name (TODO)</div>
    {/if}
  {/if}
</div>
