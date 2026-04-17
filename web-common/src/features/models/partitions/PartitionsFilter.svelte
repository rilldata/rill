<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  export let selectedFilter: string;
  export let onChange: (value: string) => void;

  let open = false;

  const options = [
    { value: "all", label: "All partitions" },
    { value: "pending", label: "Pending" },
    { value: "errors", label: "Errored" },
  ];

  $: selectedLabel =
    options.find((o) => o.value === selectedFilter)?.label ?? "All partitions";
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger
    class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {open
      ? 'bg-gray-200'
      : 'hover:bg-surface-hover'} px-2 py-1"
  >
    <span class="text-fg-secondary font-medium">{selectedLabel}</span>
    {#if open}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end" class="w-48">
    <DropdownMenu.RadioGroup
      value={selectedFilter}
      onValueChange={(v) => {
        if (v) onChange(v);
      }}
    >
      {#each options as option (option.value)}
        <DropdownMenu.RadioItem value={option.value}>
          {option.label}
        </DropdownMenu.RadioItem>
      {/each}
    </DropdownMenu.RadioGroup>
  </DropdownMenu.Content>
</DropdownMenu.Root>
