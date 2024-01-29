<script>
  import { Menu } from "@rgossiaux/svelte-headlessui";

  export let detach = false;
</script>

<Menu class={detach ? "" : "relative"}>
  <slot />
</Menu>
