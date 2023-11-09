<script lang="ts">
  import { TabGroup } from "@rgossiaux/svelte-headlessui";

  export let defaultIndex = 0;
</script>

<TabGroup {defaultIndex} on:change>
  <slot />
</TabGroup>
