<script lang="ts">
  import { MenuButton } from "@rgossiaux/svelte-headlessui";

  export let className = "";
</script>

<MenuButton class={className}><slot /></MenuButton>
