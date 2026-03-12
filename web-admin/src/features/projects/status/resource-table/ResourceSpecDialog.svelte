<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    resourceIconMapping,
    resourceLabelMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createEventDispatcher } from "svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import CanvasDescribe from "./describe/CanvasDescribe.svelte";
  import ComponentDescribe from "./describe/ComponentDescribe.svelte";
  import ConnectorDescribe from "./describe/ConnectorDescribe.svelte";
  import ExploreDescribe from "./describe/ExploreDescribe.svelte";
  import FallbackDescribe from "./describe/FallbackDescribe.svelte";
  import MetricsViewDescribe from "./describe/MetricsViewDescribe.svelte";
  import SourceModelDescribe from "./describe/SourceModelDescribe.svelte";

  export let open = false;
  export let resourceName = "";
  export let resourceKind = "";
  export let resource: V1Resource | undefined = undefined;

  // Track parent resource for back-navigation (e.g. canvas -> component)
  export let parentResourceName = "";
  export let parentResourceKind = "";
  export let parentResource: V1Resource | undefined = undefined;

  const dispatch = createEventDispatcher<{
    "view-component": { componentName: string };
    back: void;
  }>();

  $: hasParent = !!parentResource;

  $: kind = resourceKind as ResourceKind;
  $: icon = resourceIconMapping[kind];
  $: label = resourceLabelMapping[kind];
  $: filePath = resource?.meta?.filePaths?.[0];

  // Extract display name and description from resource
  $: displayName =
    kind === ResourceKind.MetricsView
      ? resource?.metricsView?.spec?.displayName
      : kind === ResourceKind.Explore
        ? resource?.explore?.spec?.displayName
        : kind === ResourceKind.Canvas
          ? resource?.canvas?.spec?.displayName
          : undefined;
  $: description =
    kind === ResourceKind.MetricsView
      ? resource?.metricsView?.spec?.description
      : kind === ResourceKind.Explore
        ? resource?.explore?.spec?.description
        : kind === ResourceKind.Component
          ? resource?.component?.spec?.description
          : undefined;
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <!-- Header: icon + type > name -->
    <Dialog.Header>
      {#if hasParent}
        <button
          class="flex items-center gap-x-1 text-xs text-fg-muted hover:text-fg-secondary mb-1 transition-colors"
          on:click={() => dispatch("back")}
        >
          <span>&larr;</span>
          <span>Back to {resourceLabelMapping[parentResourceKind] ?? parentResourceKind}</span>
        </button>
      {/if}
      <div class="flex items-center gap-x-2">
        {#if icon}
          <svelte:component this={icon} size="16px" />
        {/if}
        <Dialog.Title class="flex items-center gap-x-1.5">
          {#if label}
            <span class="text-fg-secondary">{label}</span>
            <span class="text-fg-muted">&rsaquo;</span>
          {/if}
          <span>{displayName || resourceName}</span>
        </Dialog.Title>
      </div>
      {#if description}
        <p class="text-xs text-fg-secondary mt-1">{description}</p>
      {/if}
      {#if filePath}
        <p class="text-xs text-fg-muted font-mono mt-1">
          {removeLeadingSlash(filePath)}
        </p>
      {/if}
    </Dialog.Header>

    <hr class="border-border -mx-6 mb-2" />

    <!-- Body: per-type content -->
    <div class="overflow-auto flex-1 min-h-0">
      {#if !resource}
        <p class="text-sm text-fg-secondary">No resource data available</p>
      {:else if kind === ResourceKind.Connector && resource.connector}
        <ConnectorDescribe connector={resource.connector} />
      {:else if kind === ResourceKind.Source && resource.source}
        <SourceModelDescribe source={resource.source} />
      {:else if kind === ResourceKind.Model && resource.model}
        <SourceModelDescribe model={resource.model} />
      {:else if kind === ResourceKind.MetricsView && resource.metricsView}
        <MetricsViewDescribe metricsView={resource.metricsView} />
      {:else if kind === ResourceKind.Explore && resource.explore}
        <ExploreDescribe explore={resource.explore} />
      {:else if kind === ResourceKind.Component && resource.component}
        <ComponentDescribe component={resource.component} />
      {:else if kind === ResourceKind.Canvas && resource.canvas}
        <CanvasDescribe
          canvas={resource.canvas}
          on:view-component={(e) => dispatch("view-component", e.detail)}
        />
      {:else}
        <FallbackDescribe {resource} />
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>
