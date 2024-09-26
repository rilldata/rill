<script lang="ts">
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { Button } from "../../../components/button";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { featureFlags } from "../../feature-flags";
  import ViewAsButton from "../granular-access-policies/ViewAsButton.svelte";
  import { useDashboardPolicyCheck } from "../granular-access-policies/useDashboardPolicyCheck";
  import DeployDashboardCta from "./DeployDashboardCTA.svelte";

  export let exploreName: string;

  $: exploreQuery = useExplore($runtime.instanceId, exploreName);
  $: filePath = $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "";

  $: dashboardPolicyCheck = useDashboardPolicyCheck(
    $runtime.instanceId,
    filePath,
  );

  const { readOnly } = featureFlags;

  // TODO
  // function fireTelemetry() {
  //   behaviourEvent
  //     .fireNavigationEvent(
  //       exploreName,
  //       BehaviourEventMedium.Button,
  //       MetricsEventSpace.Workspace,
  //       MetricsEventScreenName.Dashboard,
  //       MetricsEventScreenName.MetricsDefinition,
  //     )
  //     .catch(console.error);
  // }
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if $dashboardPolicyCheck.data}
    <ViewAsButton />
  {/if}
  <GlobalDimensionSearch />
  {#if !$readOnly}
    <Button href={`/files${filePath}`} type="secondary">
      Edit Metrics <MetricsIcon size="16px" />
    </Button>
    <DeployDashboardCta />
    <LocalAvatarButton />
  {/if}
</div>
