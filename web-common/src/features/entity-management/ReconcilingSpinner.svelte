<script lang="ts">
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getReconcilingItems } from "@rilldata/web-common/features/entity-management/resources-store";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: reconcilingItems = getReconcilingItems();

  const KindToName: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.Source]: "source",
    [ResourceKind.Model]: "model",
    [ResourceKind.MetricsView]: "dashboard",
  };
</script>

<div class="h-full flex flex-col gap-y-2 items-center justify-center">
  <Spinner size="1.5em" status={EntityStatus.Running} />
  <div class="flex flex-col gap-y-1">
    {#each $reconcilingItems as reconcilingItem}
      <div>
        Ingesting {KindToName[reconcilingItem.kind]}
        <span class="font-mono font-medium">{reconcilingItem.name}</span>
      </div>
    {/each}
  </div>
</div>
