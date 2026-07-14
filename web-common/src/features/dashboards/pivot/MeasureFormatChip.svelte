<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { Palette } from "lucide-svelte";
  import MeasureFormattingControls from "./MeasureFormattingControls.svelte";
  import PivotChip from "./PivotChip.svelte";
  import type { PivotChipData, PivotMeasureFormatting } from "./types";

  type Props = {
    item: PivotChipData;
    removable?: boolean;
    grab?: boolean;
    fullWidth?: boolean;
    fmt: PivotMeasureFormatting | undefined;
    lowerIsBetter?: boolean;
    onFormatChange: (fmt: PivotMeasureFormatting | null) => void;
    onRemove?: () => void;
    onmousedown?: (e: MouseEvent) => void;
  };

  let {
    item,
    removable = false,
    grab = false,
    fullWidth = false,
    fmt,
    lowerIsBetter = false,
    onFormatChange,
    onRemove = () => {},
    onmousedown = undefined,
  }: Props = $props();

  let open = $state(false);
</script>

<Popover bind:open>
  <PopoverTrigger>
    {#snippet child({ props })}
      <div {...props}>
        <PivotChip
          {item}
          {removable}
          {grab}
          {fullWidth}
          {onmousedown}
          {onRemove}
        >
          <div class="format-dropdown flex items-center gap-x-1" slot="body">
            {#if fmt}
              <Palette size="12px" />
            {/if}
            <span
              class={["flex-none transition-transform", open && "-rotate-180"]}
            >
              <CaretDownIcon size="12px" />
            </span>
          </div>
        </PivotChip>
      </div>
    {/snippet}
  </PopoverTrigger>
  <PopoverContent align="start" side="bottom" class="w-[340px] p-0">
    <div class="flex flex-col gap-y-0.5 px-3.5 pb-2.5 pt-3">
      <span class="text-sm font-semibold truncate" title={item.title}>
        {item.title}
      </span>
      <span class="text-xs text-fg-secondary">
        Conditional formatting{#if fmt?.mode === "rules"}
          · {fmt.rules.length}
          {fmt.rules.length === 1 ? "rule" : "rules"}{/if}
      </span>
    </div>
    <hr class="border-gray-200" />
    <div class="px-3.5 pb-3 pt-2.5">
      <MeasureFormattingControls
        id={item.id}
        {fmt}
        {lowerIsBetter}
        onChange={onFormatChange}
      />
    </div>
  </PopoverContent>
</Popover>
