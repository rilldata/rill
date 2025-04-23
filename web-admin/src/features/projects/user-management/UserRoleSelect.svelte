<script lang="ts">
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let value: string;

  let open = false;

  const Options = [
    {
      value: "admin",
      label: "Admin",
      description: "Full access to org settings, members, and all projects",
    },
    {
      value: "editor",
      label: "Editor",
      description: "Can create/manage projects and non-admin members",
    },
    {
      value: "viewer",
      label: "Viewer",
      description: "Read-only access to all org projects",
    },
    {
      value: "guest",
      label: "Guest",
      description: "Access to invited projects only",
    },
  ];
  function onSelect(val: string) {
    value = val;
  }
  $: selected = Options.find((o) => o.value === value);
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm {open
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    <div class="text-xs">{selected?.label ?? ""}</div>
    {#if open}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenuTrigger>
  <DropdownMenuContent side="bottom" align="end">
    {#each Options as { value, label, description } (value)}
      <DropdownMenuItem
        on:click={() => onSelect(value)}
        class="text-xs hover:bg-slate-100 {selected?.value === value
          ? 'bg-slate-50'
          : ''}"
      >
        <div class="flex flex-col">
          <div class="font-medium">{label}</div>
          <div class="text-slate-500 text-[10px]">{description}</div>
        </div>
      </DropdownMenuItem>
    {/each}
  </DropdownMenuContent>
</DropdownMenu>
