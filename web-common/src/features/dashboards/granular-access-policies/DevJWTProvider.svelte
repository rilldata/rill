<script lang="ts">
  import { appScreen } from "../../../layout/app-store";
  import { MetricsEventScreenName } from "../../../metrics/service/MetricsTypes";
  import { selectedMockUserStore } from "./stores";
  import { useDevJWT } from "./useDevJWT";

  $: isDashboardPage = $appScreen?.type === MetricsEventScreenName.Dashboard;
  $: isMockUserSelected = $selectedMockUserStore !== null;
  $: devJWT = useDevJWT($selectedMockUserStore);
</script>

<slot jwt={isDashboardPage && isMockUserSelected ? $devJWT?.data?.jwt : null} />
