<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { themeControl } from "./theme-control";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  $: ({ preference } = themeControl);

  $: preferenceValue = $preference;
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger class="">Theme</DropdownMenu.SubTrigger>
  <DropdownMenu.SubContent>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={preferenceValue === "light"}
      on:click={themeControl.set.light}
    >
      Light
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={preferenceValue === "dark"}
      on:click={themeControl.set.dark}
    >
      <Tag text="Beta" height={16} />
      Dark
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={preferenceValue === "system"}
      on:click={themeControl.set.system}
    >
      System
    </DropdownMenu.CheckboxItem>
  </DropdownMenu.SubContent>
</DropdownMenu.Sub>
