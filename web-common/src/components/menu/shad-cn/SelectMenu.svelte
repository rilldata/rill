<script lang="ts">
  import type { SelectMenuItem } from "../types";
  import Button from "../../button/Button.svelte";
  import CaretDownIcon from "../../icons/CaretDownIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import SelectMenuContent from "./SelectMenuContent.svelte";

  export let options: SelectMenuItem[];
  export let selections: Array<string | number>;
  // this is fixed text that will always be displayed in the button
  export let fixedText = "";
  export let disabled = false;
  export let active = false;
  export let ariaLabel: string | null = null;

  $: firstSelectedKey = selections?.[0] ?? null;

  $: firstSelection = firstSelectedKey
    ? options.find((option) => option.key === firstSelectedKey)
    : null;
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger asChild let:builder>
    <Button builders={[builder]} type="text" aria-label={ariaLabel} on:click>
      {fixedText}
      <span class="font-bold truncate">{firstSelection?.main}</span>
      <span
        class="transition-transform"
        class:hidden={disabled}
        class:-rotate-180={active}
      >
        <CaretDownIcon />
      </span>
    </Button>
  </DropdownMenu.Trigger>

  <SelectMenuContent {options} {selections} on:select />
</DropdownMenu.Root>
