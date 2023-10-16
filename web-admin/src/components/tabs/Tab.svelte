<script lang="ts">
  import { Tab } from "@rgossiaux/svelte-headlessui";

  const baseClasses =
    "p-2 text-gray-500 text-xs font-medium hover:bg-slate-100 hover:text-gray-700 rounded flex items-center justify-center gap-x-1";
</script>

<Tab
  class={({ selected }) => {
    return `${baseClasses} ${selected ? "text-gray-900" : "text-gray-500"}`;
  }}
>
  <slot />
</Tab>
