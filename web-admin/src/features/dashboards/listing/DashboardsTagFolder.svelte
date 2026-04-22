<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretRightFilledIcon from "@rilldata/web-common/components/icons/CaretRightFilledIcon.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";

  export let tag: string;
  export let resources: V1Resource[];
  export let organization: string;
  export let project: string;
  export let isEmbedded = false;

  let open = true;

  $: count = resources.length;
</script>

<div>
  <button
    class="flex items-center gap-x-1.5 w-full px-4 py-2 bg-surface-background hover:bg-surface-hover text-left"
    onclick={() => (open = !open)}
    aria-expanded={open}
  >
    {#if open}
      <CaretDownIcon size="12px" />
    {:else}
      <CaretRightFilledIcon size="12px" />
    {/if}
    <span class="text-sm font-semibold text-fg-secondary">{tag}</span>
    <span class="text-xs text-fg-tertiary ml-1">({count})</span>
  </button>

  {#if open}
    <ul role="list" class="list-none p-0 m-0 w-full">
      {#each resources as resource (resource.meta?.name?.name)}
        {@const name = resource.meta?.name?.name ?? ""}
        {@const isMetricsExplorer = !!resource.explore}
        {@const title = isMetricsExplorer
          ? resource.explore?.spec?.displayName
          : resource.canvas?.spec?.displayName}
        {@const description = isMetricsExplorer
          ? (resource.explore?.spec?.description ?? "")
          : ""}
        {@const refreshedOn = isMetricsExplorer
          ? resource.explore?.state?.dataRefreshedOn
          : resource.canvas?.state?.dataRefreshedOn}
        <li
          class="block w-full h-[60px] border-t bg-surface-background hover:bg-surface-hover"
        >
          <DashboardsTableCompositeCell
            {name}
            title={title ?? name}
            lastRefreshed={refreshedOn ?? ""}
            {description}
            error={resource.meta?.reconcileError ?? ""}
            {isMetricsExplorer}
            {isEmbedded}
            {organization}
            {project}
            activeTag={tag}
          />
        </li>
      {/each}
    </ul>
  {/if}
</div>
