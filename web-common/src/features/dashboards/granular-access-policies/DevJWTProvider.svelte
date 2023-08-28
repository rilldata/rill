<script lang="ts">
  import { page } from "$app/stores";
  import { selectedMockUserStore } from "./stores";
  import { useDevJWT } from "./useDevJWT";

  $: devJWT = useDevJWT($selectedMockUserStore);

  $: isMockUserSelected = $selectedMockUserStore !== null;
  $: isDashboardPage = $page.route.id === "/(application)/dashboard/[name]";
</script>

<slot jwt={isMockUserSelected && isDashboardPage ? $devJWT?.data?.jwt : null} />
