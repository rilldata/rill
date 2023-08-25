<script lang="ts">
  import { selectedMockUserStore } from "./stores";
  import { useDevJWT } from "./useDevJWT";

  $: devJWT = useDevJWT($selectedMockUserStore);
</script>

<slot jwt={$selectedMockUserStore ? $devJWT?.data?.jwt : null} />
