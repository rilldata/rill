<script lang="ts">
  import { page } from "$app/stores";
  import { selectedMockUserStore } from "./stores";
  import { useDevJWT } from "./useDevJWT";

  $: isDashboardPage = $page.route.id === "/(application)/dashboard/[name]";
  $: isMockUserSelected = $selectedMockUserStore !== null;
  $: devJWT = useDevJWT($selectedMockUserStore);
</script>

<slot jwt={isDashboardPage && isMockUserSelected ? $devJWT?.data?.jwt : null} />
