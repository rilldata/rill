<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: reconcilingItems = fileArtifacts.getReconcilingResourceNames();
  $: typedReconcilingItems = $reconcilingItems.map((reconcilingItem) => ({
    name: reconcilingItem.name,
    kind: reconcilingItem.kind,
  }));

  const userFriendlyPhrases: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.API]: "Building API",
    [ResourceKind.Alert]: "Building alert",
    [ResourceKind.Canvas]: "Building canvas",
    [ResourceKind.Component]: "Building chart",
    [ResourceKind.Connector]: "Building connector",
    [ResourceKind.Explore]: "Building explore",
    [ResourceKind.MetricsView]: "Building metrics view",
    [ResourceKind.Model]: "Building model",
    [ResourceKind.Report]: "Building report",
    [ResourceKind.Source]: "Ingesting source",
    [ResourceKind.Theme]: "Building theme",
  };
</script>

<div class="size-full p-2 flex flex-col gap-y-2 items-center justify-center">
  <Spinner size="1.5em" status={EntityStatus.Running} />
  <div class="flex flex-col w-full text-center gap-y-1">
    {#each typedReconcilingItems as { name, kind }, i (i)}
      {#if name && kind}
        <div class="w-full truncate">
          {userFriendlyPhrases[kind]}

          <span class="font-mono font-medium">{name}</span>
        </div>
      {/if}
    {/each}
  </div>
</div>
