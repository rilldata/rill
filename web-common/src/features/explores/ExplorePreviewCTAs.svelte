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
  import DeployDashboardCta from "../dashboards/workspace/DeployDashboardCTA.svelte";
  import { featureFlags } from "../feature-flags";

  export let exploreName: string;

  $: exploreQuery = useExplore($runtime.instanceId, exploreName);
  $: exploreFilePath = $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "";
  $: metricsViewFilePath =
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "";

  $: dashboardPolicyCheck = useDashboardPolicyCheck(
    $runtime.instanceId,
    exploreFilePath,
  );

  const { readOnly } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if $dashboardPolicyCheck.data}
    <ViewAsButton />
  {/if}
  <GlobalDimensionSearch />
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
          <MetricsViewIcon size="16px" />Metrics View
        </DropdownMenu.Item>
        <DropdownMenu.Item href={`/files${exploreFilePath}`}>
          <ExploreIcon size="16px" />Explore
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
    <DeployDashboardCta />
  {/if}
</div>
