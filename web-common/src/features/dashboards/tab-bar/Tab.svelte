<script lang="ts">
  import { Tab } from "@rgossiaux/svelte-headlessui";

  const baseClasses = "font-semibold py-2 px-1 items-center"
</script>

<Tab
  class={({ selected }) => {
    return `${baseClasses} ${selected ? "text-blue-600 border-b-2 border-blue-500" : "text-gray-500"}`;
  }}
>
  <slot />
</Tab>
