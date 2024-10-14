<script lang="ts">
  import * as Select from "@rilldata/web-common/components/select";
  import type { Selected } from "bits-ui";
  import Button from "../../../components/button/Button.svelte";

  export let selectedFilter: string;
  export let onChange: (selected: Selected<string>) => void;

  let openFilterMenu = false;

  const options = [
    { value: "all", label: "all" },
    { value: "pending", label: "pending" },
    { value: "errors", label: "errors" },
  ];
</script>

<Select.Root
  items={options}
  onSelectedChange={onChange}
  bind:open={openFilterMenu}
>
  <Select.Trigger class="outline-none border-none w-fit px-0 gap-x-0.5">
    <Button type="text" label="Filter splits">
      <span class="text-gray-700 hover:text-inherit">
        Showing <b>{selectedFilter}</b>
      </span>
    </Button>
  </Select.Trigger>
  <Select.Content sameWidth={false} align="end">
    {#each options as option (option.value)}
      <Select.Item
        value={option.value}
        label={option.label}
        class={`text-xs flex items-start ${
          selectedFilter === option.value ? "font-bold" : ""
        }`}
      />
    {/each}
  </Select.Content>
</Select.Root>
