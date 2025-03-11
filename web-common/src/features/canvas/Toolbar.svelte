<script lang="ts">
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Copy } from "lucide-svelte";

  export let dropdownOpen = false;
  export let onDelete: () => void;
  export let onDuplicate: () => void;
</script>

<div
  class:!flex={dropdownOpen}
  class="group-hover:flex p-0 overflow-hidden bg-slate-50 gap-x-1 items-center justify-center hidden toolbar top-0 right-0 shadow-sm z-[1000] absolute w-fit border-l border-b pointer-events-auto rounded-bl-sm rounded-tr-sm"
>
  <DropdownMenu.Root bind:open={dropdownOpen}>
    <DropdownMenu.Trigger
      class="size-7 grid place-content-center hover:bg-slate-100 active:bg-slate-200"
    >
      <ThreeDot size="16px" />
    </DropdownMenu.Trigger>

    <DropdownMenu.Content
      align="end"
      sideOffset={8}
      alignOffset={-4}
      class="w-40"
    >
      <DropdownMenu.Item on:click={onDuplicate}>
        <Copy size="14px" />
        Duplicate
      </DropdownMenu.Item>
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        on:click={onDelete}
        class="text-red-600 data-[highlighted]:text-red-600"
      >
        <Trash size="14px" />
        Delete
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>
