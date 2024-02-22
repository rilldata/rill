<script lang="ts">
  import { TabList } from "@rgossiaux/svelte-headlessui";
</script>

<TabList class="flex gap-x-6" let:selectedIndex>
  <slot {selectedIndex} />
</TabList>
