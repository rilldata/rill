<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getConnectorsWithMetricsViewSupport } from "@rilldata/web-common/features/entity-management/add/selectors.ts";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import {
    connectorClassMapping,
    connectorIconMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import { createResourceAndNavigate } from "@rilldata/web-common/features/entity-management/add/new-files.ts";
  import File from "@rilldata/web-common/components/icons/File.svelte";

  export let onSelect: (connector: string) => void;

  const runtimeClient = useRuntimeClient();

  const connectorsWithMetricsViewSupportQuery =
    getConnectorsWithMetricsViewSupport(runtimeClient);
  $: connectorsWithMetricsViewSupport =
    $connectorsWithMetricsViewSupportQuery.data ?? [];
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger class="flex gap-x-2" aria-label="Add metrics view">
    <svelte:component
      this={resourceIconMapping[ResourceKind.MetricsView]}
      size="16px"
    />
    Metrics view
  </DropdownMenu.SubTrigger>
  <DropdownMenu.SubContent class="w-[240px]">
    {#each connectorsWithMetricsViewSupport as { displayName, connector, schema } (connector)}
      {@const icon = connectorIconMapping[schema]}
      {@const className = connectorClassMapping[schema] ?? ""}
      <DropdownMenu.Item
        on:click={() => onSelect(connector)}
        class="flex gap-x-2"
        aria-label="Create metrics view for {displayName}"
      >
        <svelte:component this={icon} size="16px" class={className} />
        {displayName}
      </DropdownMenu.Item>
    {/each}
    <DropdownMenu.Item
      class="flex gap-x-2"
      on:click={() =>
        createResourceAndNavigate(runtimeClient, ResourceKind.MetricsView)}
      aria-label="Blank metrics view"
    >
      <File size="14px" className="stroke-icon-muted" />
      Blank metrics view
    </DropdownMenu.Item>
  </DropdownMenu.SubContent>
</DropdownMenu.Sub>
