<script lang="ts">
  import Button from "../Button.svelte";

  export let primary = false;
</script>

<Button on:click {primary}>
  <slot />
</Button>
