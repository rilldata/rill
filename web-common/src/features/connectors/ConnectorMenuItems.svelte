<script lang="ts">
  import { Settings } from "lucide-svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { addSourceModal } from "@rilldata/web-common/features/sources/modal/add-source-visibility";
  import { getSchemaNameFromDriver } from "@rilldata/web-common/features/sources/modal/connector-schemas";
  import { AI_CONNECTORS } from "@rilldata/web-common/features/sources/modal/constants";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";

  export let filePath: string;

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher<{ openAiDialog: void }>();

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: connectorName = $resourceQuery.data?.meta?.name?.name;
  $: driverName = $resourceQuery.data?.connector?.spec?.driver;
  $: spec = $resourceQuery.data?.connector?.spec;
  $: schemaName = driverName ? getSchemaNameFromDriver(driverName) : null;

  // Hide edit for managed connectors (provisioned ClickHouse, DuckDB)
  $: isManaged = spec?.provision === true || driverName === "duckdb";
  $: isAiConnector = schemaName ? AI_CONNECTORS.includes(schemaName) : false;

  function editConnector() {
    if (!schemaName || !connectorName) return;

    if (isAiConnector) {
      dispatch("openAiDialog");
      return;
    }

    addSourceModal.openForEdit(schemaName, connectorName);
  }
</script>

{#if !isManaged}
  <NavigationMenuItem on:click={editConnector}>
    <Settings slot="icon" size="12px" />
    Edit Connector
  </NavigationMenuItem>
{/if}
