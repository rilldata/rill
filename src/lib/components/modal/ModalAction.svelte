<script lang="ts">
  import { Button } from "$lib/components/button";

  export let primary = false;
</script>

<Button on:click type={primary ? "primary" : "secondary"}>
  <slot />
</Button>
