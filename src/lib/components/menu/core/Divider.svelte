<script lang="ts">
  import { getContext } from "svelte";
  export let marginTop = 2;
  export let marginBottom = 2;
  const dark = getContext("rill:menu:dark");
  // mt-2 mb-2 mt-1 mb-1
</script>

<hr
  class="{dark
    ? 'border-gray-500'
    : 'border-gray-200'} mt-{marginTop} mb-{marginBottom}"
/>
