<script lang="ts">
  import { appScreen } from "../../../layout/app-store";
  import { MetricsEventScreenName } from "../../../metrics/service/MetricsTypes";
  import { selectedMockUserJWT } from "./stores";

  $: isDashboardPage = $appScreen?.type === MetricsEventScreenName.Dashboard;
  $: isMockUserSelected = $selectedMockUserJWT !== null;
  $: devJWT = $selectedMockUserJWT;
</script>

<slot jwt={isDashboardPage && isMockUserSelected ? devJWT : null} />
