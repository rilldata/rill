<script lang="ts">
  import {
    connectors,
    getConnectorSchema,
    isMultiStepConnector,
    toConnectorDriver,
  } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import AddDataForm from "@rilldata/web-common/features/sources/modal/AddDataForm.svelte";
  import NewSourceSelector from "@rilldata/web-common/features/welcome/new-sources/NewSourceSelector.svelte";

  export let selectedConnectorName: string | null = null;

  $: selectedConnectorInfo = connectors.find(
    (c) => c.name === selectedConnectorName,
  );
  $: selectedConnectorDriver = selectedConnectorInfo
    ? toConnectorDriver(selectedConnectorInfo)
    : null;

  $: isConnectorType =
    selectedConnectorDriver?.implementsObjectStore ||
    selectedConnectorDriver?.implementsOlap ||
    selectedConnectorDriver?.implementsSqlStore ||
    (selectedConnectorDriver?.implementsWarehouse &&
      selectedConnectorDriver?.name !== "salesforce") ||
    isMultiStepConnector(
      getConnectorSchema(
        selectedConnectorName ?? selectedConnectorDriver?.name ?? "",
      ),
    );
</script>

{#if selectedConnectorName && selectedConnectorDriver}
  <AddDataForm
    connector={selectedConnectorDriver}
    schemaName={selectedConnectorName}
    formType={isConnectorType ? "connector" : "source"}
    onClose={() => {}}
    onCloseAfterNavigation={() => {}}
    onBack={() => {
      selectedConnectorName = null;
    }}
  />
{:else}
  <NewSourceSelector
    startConnectorSelection={(name) => {
      selectedConnectorName = name;
    }}
  />
{/if}
