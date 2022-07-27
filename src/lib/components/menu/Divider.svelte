<script lang="ts">
  import { getContext } from "svelte";

  const dark = getContext("rill:menu:dark");
</script>

<hr class="{dark ? 'border-gray-500' : 'border-gray-200'} mt-2 mb-2" />
