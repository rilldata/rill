<script lang="ts">
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import type {
    V1MetricsViewSpec,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import { isErrorStoreEmpty } from "../errors/error-store";

  export let instanceId: string;
  export let activeResourceName: string;

  const dispatch = createEventDispatcher();

  // Project breadcrumb (if any)
  $: onProjectPage = !activeResourceName;

  // Dashboard breadcrumb
  $: dashboards = useValidDashboards(instanceId);
  let currentResource: V1Resource;
  $: currentResource = $dashboards?.data?.find(
    (listing) => listing.meta.name.name === activeResourceName,
  );
  $: currentDashboardName = currentResource?.meta?.name?.name;
  let currentDashboard: V1MetricsViewSpec;
  $: currentDashboard = currentResource?.metricsView?.state?.validSpec;
  $: onDashboardPage = !!activeResourceName;
</script>

<div class="flex items-center w-full pr-4 {onProjectPage ? '' : 'border-b'}">
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
        {#if currentDashboard}
          <BreadcrumbItem
            options={$dashboards.data.reduce((map, { meta, metricsView }) => {
              const id = meta.name.name;
              return map.set(id, {
                label: metricsView?.state?.validSpec?.title || id,
              });
            }, new Map())}
            current={currentDashboardName}
            onSelect={(dashboard) => dispatch("select-dashboard", dashboard)}
            isCurrentPage={onDashboardPage}
          />
        {/if}
      </ol>
    </nav>
  {:else}
    <div />
  {/if}
  {#if onDashboardPage}
    <div class="grow" />
    <div class="flex gap-x-4 items-center">
      <LastRefreshedDate dashboard={activeResourceName} />
    </div>
  {/if}
</div>
