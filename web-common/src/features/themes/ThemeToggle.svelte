<script lang="ts">
  import { MoonIcon, SunIcon } from "lucide-svelte";
  import { themeControl } from "./theme-control";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  $: mode = $themeControl;
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    class="hover:bg-gray-100 size-full border aspect-square rounded-md grid place-content-center text-gray-800"
  >
    {#if mode === "dark"}
      <MoonIcon size="14px" />
    {:else}
      <SunIcon size="14px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    <DropdownMenu.Item on:click={themeControl.set.light}>
      Light
    </DropdownMenu.Item>
    <DropdownMenu.Item on:click={themeControl.set.dark}>Dark</DropdownMenu.Item>
    <DropdownMenu.Item on:click={themeControl.set.system}>
      System
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
