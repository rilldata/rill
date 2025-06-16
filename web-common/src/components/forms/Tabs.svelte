<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import {
    Tabs,
    TabsList,
    TabsTrigger,
    TabsContent,
  } from "@rilldata/web-common/components/tabs";

  export let value: string;
  export let options: { value: string; label: string }[];
  const dispatch = createEventDispatcher();

  function handleChange(e: CustomEvent) {
    dispatch("change", e.detail);
  }
</script>

<Tabs {value} on:change={handleChange} class="w-full">
  <TabsList
    class="bg-muted h-9 rounded-[10px] w-full border border-gray-200 flex"
  >
    {#each options as opt}
      <TabsTrigger
        value={opt.value}
        class="flex-1 h-7 rounded-[8px] py-2 border border-gray-200 text-sm text-[#09090B] transition-all relative z-0
          data-[state=active]:bg-white data-[state=active]:shadow-lg  data-[state=active]:z-10 data-[state=active]:outline-2 data-[state=active]:outline-gray-200
          data-[state=inactive]:bg-transparent data-[state=inactive]:z-0 data-[state=inactive]:outline-none data-[state=inactive]:border-none
          focus-visible:outline-none"
        style="box-shadow: none;"
      >
        {opt.label}
      </TabsTrigger>
    {/each}
  </TabsList>
  <slot />
</Tabs>
