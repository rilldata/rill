<script lang="ts">
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import { useValidVisualizations } from "@rilldata/web-common/features/dashboards/selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type {
    V1Resource,
    V1ResourceName,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import { isErrorStoreEmpty } from "../errors/error-store";

  export let instanceId: string;
  export let activeResource: V1ResourceName;

  const dispatch = createEventDispatcher();

  $: onProjectPage = !activeResource;
  $: onMetricsExplorerPage =
    !!activeResource &&
    activeResource.kind === ResourceKind.MetricsView.toString();
  // $: onCustomDashboardPage =
  //   !!activeResourceName &&
  //   activeResourceKind === ResourceKind.Dashboard.toString();

  // Dashboard breadcrumb
  $: visualizationsQuery = useValidVisualizations(instanceId);
  $: ({ data: visualizations } = $visualizationsQuery);
  let currentResource: V1Resource;
  $: currentResource = visualizations?.find(
    (listing) => listing.meta.name.name === activeResource?.name,
  );
  $: currentResourceName = currentResource?.meta?.name?.name;

  $: breadcrumbOptions = visualizations?.reduce(
    (map, { meta, metricsView, dashboard }) => {
      const name = meta.name.name;
      const isMetricsExplorer = !!metricsView;
      return map.set(name.toLowerCase(), {
        label:
          (isMetricsExplorer
            ? metricsView?.state?.validSpec?.title
            : dashboard?.spec?.title) || name,
      });
    },
    new Map(),
  );

  function onSelectResource(name: string) {
    // Because the breadcrumb only returns the identifying name, we need to look up the V1ResourceName (name + kind)
    const resource = visualizations?.find(
      (listing) => listing.meta.name.name === name,
    );
    dispatch("select-resource", resource.meta.name);
  }
</script>

<div class="flex items-center w-full pr-4 py-1" class:border-b={!onProjectPage}>
  {#if $isErrorStoreEmpty}
    <nav>
      <ol class="flex items-center pl-4">
        {#if !onProjectPage}
          <div class="flex gap-x-2">
            <button
              class="text-gray-500 hover:text-gray-600"
              on:click={() => dispatch("go-home")}>Home</button
            >
            <span class="text-gray-600">/</span>
          </div>
        {/if}

        {#if currentResource}
          <BreadcrumbItem
            options={breadcrumbOptions}
            current={currentResourceName}
            onSelect={onSelectResource}
            isCurrentPage
            isEmbedded
          />
        {/if}
      </ol>
    </nav>
  {:else}
    <div />
  {/if}

  {#if onMetricsExplorerPage}
    <div class="grow" />
    <div class="flex gap-x-4 items-center">
      <LastRefreshedDate dashboard={activeResource?.name} />
    </div>
  {/if}
</div>
