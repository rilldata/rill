<script lang="ts">
  import { invalidateAllMetricsViews } from "@rilldata/web-common/runtime-client/invalidation";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { appScreen } from "../../../layout/app-store";
  import { MetricsEventScreenName } from "../../../metrics/service/MetricsTypes";
  import { selectedMockUserStore } from "./stores";
  import { useDevJWT } from "./useDevJWT";

  export let instanceId: string;

  $: isDashboardPage = $appScreen?.type === MetricsEventScreenName.Dashboard;
  $: isMockUserSelected = $selectedMockUserStore !== null;
  $: devJWT = useDevJWT($selectedMockUserStore);

  const queryClient = useQueryClient();

  // TODO: this is temporary fix. We should avoid global reactive statement to invalidate queries.
  //       perhaps move this to the place where we actually change the jwt in ViewAsButton.svelte
  $: (devJWT || devJWT === null) &&
    invalidateAllMetricsViews(queryClient, instanceId);
</script>

<slot jwt={isDashboardPage && isMockUserSelected ? $devJWT?.data?.jwt : null} />
