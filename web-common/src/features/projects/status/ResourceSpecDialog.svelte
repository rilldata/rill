<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    resourceIconMapping,
    resourceLabelMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import StructuredView from "./describe/StructuredView.svelte";

  let {
    open = $bindable(false),
    resourceName = "",
    resourceKind = "",
    resource = undefined,
    parentResourceKind = "",
    parentResource = undefined,
    allResources = [],
    onback,
    onviewcomponent,
  }: {
    open?: boolean;
    resourceName?: string;
    resourceKind?: string;
    resource?: V1Resource;
    parentResourceKind?: string;
    parentResource?: V1Resource;
    allResources?: V1Resource[];
    onback?: () => void;
    onviewcomponent?: (componentName: string) => void;
  } = $props();

  let hasParent = $derived(!!parentResource);

  let kind = $derived(resourceKind as ResourceKind);
  let icon = $derived(resourceIconMapping[kind]);
  let label = $derived(resourceLabelMapping[kind]);
  let filePath = $derived(resource?.meta?.filePaths?.[0]);

  function getDisplayName(
    k: ResourceKind,
    r: V1Resource | undefined,
  ): string | undefined {
    switch (k) {
      case ResourceKind.MetricsView:
        return r?.metricsView?.spec?.displayName;
      case ResourceKind.Explore:
        return r?.explore?.spec?.displayName;
      case ResourceKind.Canvas:
        return r?.canvas?.spec?.displayName;
      default:
        return undefined;
    }
  }

  let displayName = $derived(getDisplayName(kind, resource));
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <!-- Header: icon + type > name -->
    <Dialog.Header>
      {#if hasParent}
        <button
          class="flex items-center gap-x-1 text-xs text-fg-muted hover:text-fg-secondary mb-1 transition-colors"
          aria-label="Back to parent resource"
          onclick={() => onback?.()}
        >
          <span>&larr;</span>
          <span
            >Back to {resourceLabelMapping[parentResourceKind] ??
              parentResourceKind}</span
          >
        </button>
      {/if}
      <div class="flex items-center gap-x-2">
        {#if icon}
          {@const Icon = icon}
          <Icon size="16px" />
        {/if}
        <Dialog.Title class="flex items-center gap-x-1.5">
          {#if label}
            <span class="text-fg-secondary">{label}</span>
            <span class="text-fg-muted">&rsaquo;</span>
          {/if}
          <span>{displayName || resourceName}</span>
        </Dialog.Title>
      </div>
      {#if filePath}
        <p class="text-xs text-fg-muted font-mono mt-1">
          {removeLeadingSlash(filePath)}
        </p>
      {/if}
    </Dialog.Header>

    <hr class="border-border -mx-6 mb-2" />

    <!-- Body -->
    <div class="overflow-y-auto min-h-0 flex-1">
      {#if !resource}
        <p class="text-sm text-fg-secondary">No resource data available</p>
      {:else}
        <StructuredView {resource} {allResources} />
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>
