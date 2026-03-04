<script lang="ts">
  import {
    getConnectorSchema,
    isMultiStepConnector,
  } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import type { LayoutData } from "./$types";
  import { goto } from "$app/navigation";

  export let data: LayoutData;
  $: ({ connectorName, connectorDriver } = data);

  $: isConnectorType =
    connectorDriver?.implementsObjectStore ||
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    (connectorDriver?.implementsWarehouse &&
      connectorDriver?.name !== "salesforce") ||
    isMultiStepConnector(
      getConnectorSchema(connectorName ?? connectorDriver?.name ?? ""),
    );
  $: console.log(isConnectorType, connectorDriver);
</script>

{#if connectorDriver}
  <ConnectorForm
    connector={connectorDriver}
    onSubmit={() => void goto(`/welcome/sources/${connectorName}/tables`)}
    onBack={() => window.history.back()}
  />
{/if}

<style lang="postcss">
</style>
