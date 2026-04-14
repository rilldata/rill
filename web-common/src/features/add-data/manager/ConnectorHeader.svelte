<script lang="ts">
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-metadata.ts";
  import { connectorInfoMap } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { getAnalyzedConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { AddDataConfig } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import {
    getConnectorDriverForSchema,
    isConnectorType,
  } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
  import { inferSchemaForConnector } from "@rilldata/web-common/features/entity-management/add/selectors.ts";

  export let config: AddDataConfig;
  export let schemaName: string;
  export let connectorName: string | undefined = undefined;
  export let onConnectorChange: (connectorName: string) => void;
  export let onNewConnector: () => void;

  $: connectorInfo = connectorInfoMap.get(schemaName);
  $: displayIcon = connectorIconMapping[schemaName];

  const runtimeClient = useRuntimeClient();
  $: analyzedConnectorsQuery = getAnalyzedConnectors(runtimeClient, false);
  // TODO: some schema will share driver name, differentiate them.
  $: connectorsForSchema = $analyzedConnectorsQuery.data?.connectors.filter(
    (connector) => inferSchemaForConnector(connector) === schemaName,
  );
  $: connectorOptions =
    connectorsForSchema?.map((connector) => ({
      label: connector.name!,
      value: connector.name!,
    })) ?? [];

  $: driverForSchema = getConnectorDriverForSchema(schemaName);
  $: showConnectorSelector =
    connectorName &&
    driverForSchema &&
    isConnectorType(driverForSchema) &&
    !config.welcomeScreen &&
    connectorOptions.length > 0;
</script>

{#if connectorInfo}
  <div class="flex flex-row items-center px-6 py-4 gap-1 border-b">
    {#if displayIcon}
      <svelte:component this={displayIcon} size="18px" />
    {/if}
    <span class="text-lg leading-none font-semibold">
      {connectorInfo.displayName}
    </span>

    {#if showConnectorSelector}
      <Select
        id="connector"
        value={connectorName}
        options={connectorOptions}
        onChange={onConnectorChange}
        onAddNew={onNewConnector}
        addNewLabel="+ {connectorInfo.displayName} connector"
        outline={false}
        ariaLabel="Select connector"
      />
    {/if}
  </div>
{/if}
