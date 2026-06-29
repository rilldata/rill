<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { themeControl } from "./theme-control";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  $: ({ preference } = themeControl);

  $: preferenceValue = $preference;
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger class="">{m.theme_label()}</DropdownMenu.SubTrigger>
  <DropdownMenu.SubContent>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={preferenceValue === "light"}
      onclick={themeControl.set.light}
    >
      {m.theme_light()}
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={preferenceValue === "dark"}
      onclick={themeControl.set.dark}
    >
      <Tag text={m.theme_beta()} height={16} />
      {m.theme_dark()}
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={preferenceValue === "system"}
      onclick={themeControl.set.system}
    >
      {m.theme_system()}
    </DropdownMenu.CheckboxItem>
  </DropdownMenu.SubContent>
</DropdownMenu.Sub>
