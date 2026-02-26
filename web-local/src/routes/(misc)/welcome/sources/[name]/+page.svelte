<script lang="ts">
  import {
    getConnectorSchema,
    isMultiStepConnector,
  } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import AddDataForm from "@rilldata/web-common/features/sources/modal/AddDataForm.svelte";
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
</script>

{#if connectorDriver}
  <AddDataForm
    connector={connectorDriver}
    schemaName={connectorName}
    formType={isConnectorType ? "connector" : "source"}
    onClose={() => {
      void goto(`/welcome/sources/${connectorName}/tables`);
    }}
    onCloseAfterNavigation={() => {}}
    onBack={() => window.history.back()}
    isSubmitting={false}
  />
{/if}

<style lang="postcss">
</style>
