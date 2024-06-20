<script lang="ts">
  import { MenuItems } from "@rgossiaux/svelte-headlessui";

  export let positioningOverride = "";

  // Temporary approach for allowing various positioning options. See how it goes.
  let positioning = positioningOverride || "absolute left-0";
</script>

<MenuItems
  class="{positioning} z-[1000] w-fit mt-1 bg-background border border-slate-100 rounded-sm shadow outline-none max-h-60 overflow-y-auto focus:outline-none"
>
  <slot />
</MenuItems>
