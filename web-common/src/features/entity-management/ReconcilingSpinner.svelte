<script lang="ts">
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { getReconcilingItems } from "@rilldata/web-common/features/entity-management/resources-store";

  $: reconcilingItems = getReconcilingItems();

  const KindToName: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.Source]: "source",
    [ResourceKind.Model]: "model",
    [ResourceKind.MetricsView]: "dashboard",
  };
</script>

<div class="grid h-full place-content-center align-middle">
  <div class="grid grid-flow-col">
    <div class="w-8">
      <Spinner
        duration={300 + Math.random() * 200}
        size="1.5em"
        status={EntityStatus.Running}
      />
    </div>
    <div>
      {#each $reconcilingItems as reconcilingItem}
        <div>
          Ingesting {KindToName[reconcilingItem.kind]}
          "{reconcilingItem.name}"
        </div>
      {/each}
    </div>
  </div>
</div>
