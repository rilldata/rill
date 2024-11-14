<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { Button } from "../../components/button";
  import { runtime } from "../../runtime-client/runtime-store";
  import ViewAsButton from "../dashboards/granular-access-policies/ViewAsButton.svelte";
  import { useDashboardPolicyCheck } from "../dashboards/granular-access-policies/useDashboardPolicyCheck";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import { resourceColorMapping } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { featureFlags } from "../feature-flags";

  export let exploreName: string;

  $: exploreQuery = useExplore($runtime.instanceId, exploreName);
  $: exploreFilePath = $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "";
  $: metricsViewFilePath =
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "";
  $: metricsViewName = $exploreQuery.data?.metricsView?.meta?.name?.name ?? "";

  $: explorePolicyCheck = useDashboardPolicyCheck(
    $runtime.instanceId,
    exploreFilePath,
  );
  $: metricsPolicyCheck = useDashboardPolicyCheck(
    $runtime.instanceId,
    metricsViewFilePath,
  );

  const { readOnly } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if $explorePolicyCheck.data || $metricsPolicyCheck.data}
    <ViewAsButton />
  {/if}
  <StateManagersProvider {metricsViewName} {exploreName}>
    <GlobalDimensionSearch />
  </StateManagersProvider>
  {#if !$readOnly}
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild let:builder>
        <Button type="secondary" builders={[builder]}>
          Edit
          <CaretDownIcon />
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Item href={`/files${metricsViewFilePath}`}>
          <MetricsViewIcon
            color={resourceColorMapping[ResourceKind.MetricsView]}
            size="16px"
          />Metrics View
        </DropdownMenu.Item>
        <DropdownMenu.Item href={`/files${exploreFilePath}`}>
          <ExploreIcon
            color={resourceColorMapping[ResourceKind.Explore]}
            size="16px"
          />Explore
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>
