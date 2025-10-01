<script lang="ts">
  import { isErrorStoreEmpty } from "@rilldata/web-admin/components/errors/error-store";
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import TwoTieredBreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/TwoTieredBreadcrumbItem.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type {
    V1Resource,
    V1ResourceName,
  } from "@rilldata/web-common/runtime-client";

  export let instanceId: string;
  export let activeResource: V1ResourceName;

  const { twoTieredNavigation } = featureFlags;

  $: onProjectPage = !activeResource;

  // Dashboard breadcrumb
  $: dashboardsQuery = useValidDashboards(instanceId);
  $: ({ data: dashboards } = $dashboardsQuery);
  let currentResource: V1Resource;
  $: currentResource = dashboards?.find(
    (listing) => listing.meta.name.name === activeResource?.name,
  );
  $: currentResourceName = currentResource?.meta?.name?.name;

  $: breadcrumbOptions = dashboards?.reduce(
    (map, { meta, explore, canvas }) => {
      const name = meta.name.name;
      const isExplore = !!explore;
      return map.set(name.toLowerCase(), {
        label:
          (isExplore
            ? explore?.state?.validSpec?.displayName
            : canvas?.state?.validSpec?.displayName) || name,
        href: `/-/embed/${isExplore ? "explore" : "canvas"}/${name}`,
        preloadData: false,
      });
    },
    new Map(),
  );
</script>

{#if $isErrorStoreEmpty}
  <nav>
    <ol class="flex items-center pl-4">
      {#if !onProjectPage}
        <div class="flex gap-x-2">
          <a class="text-gray-500 hover:text-gray-600" href="/-/embed">
            Home
          </a>
          <span class="text-gray-600">/</span>
        </div>
      {/if}

      {#if currentResource}
        {#if $twoTieredNavigation}
          <TwoTieredBreadcrumbItem
            options={breadcrumbOptions}
            current={currentResourceName}
            isCurrentPage
          />
        {:else}
          <BreadcrumbItem
            options={breadcrumbOptions}
            current={currentResourceName}
            isCurrentPage
            isEmbedded
          />
        {/if}
      {/if}
    </ol>
  </nav>
{:else}
  <div />
{/if}
