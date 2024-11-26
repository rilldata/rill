<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { RefreshCcwIcon } from "lucide-svelte";

  export let resourceKind: string;
  export let resourceName: string;
  export let isSource: boolean;
  export let refreshSource: (
    resourceKind: string,
    resourceName: string,
  ) => void;

  let isDropdownOpen = false;
</script>

{#if isSource}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={() => refreshSource(resourceKind, resourceName)}
      >
        <RefreshCcwIcon size="12px" />
        <span class="ml-2">Refresh</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
