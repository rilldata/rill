<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import { connectorInfoMap } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { getAnalyzedConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let schemaName: string;
  // export let onConnectorChange: (connector: V1ConnectorDriver) => void;
  // export let onNewConnector: () => void;

  $: connectorInfo = connectorInfoMap.get(schemaName);
  $: displayIcon = connectorIconMapping[schemaName];

  const runtimeClient = useRuntimeClient();
  $: analyzedConnectorsQuery = getAnalyzedConnectors(runtimeClient, false);
  // TODO: some schema will share driver name, differentiate them.
  $: connectorsForSchema = $analyzedConnectorsQuery.data?.connectors.filter(
    (connector) => connector.driver?.name === schemaName,
  );
</script>

{#if connectorInfo}
  <div class="flex flex-row items-center px-6 py-4 gap-1 border-b">
    {#if displayIcon}
      <svelte:component this={displayIcon} size="18px" />
    {/if}
    <span class="text-lg leading-none font-semibold">
      {connectorInfo.displayName}
    </span>
  </div>
{/if}
