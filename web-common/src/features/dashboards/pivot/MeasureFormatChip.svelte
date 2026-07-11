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
    fmt: PivotMeasureFormatting | undefined;
    onFormatChange: (fmt: PivotMeasureFormatting | null) => void;
    onRemove?: () => void;
    onmousedown?: (e: MouseEvent) => void;
  };

  let {
    item,
    removable = false,
    grab = false,
    fmt,
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
        <PivotChip {item} {removable} {grab} {onmousedown} {onRemove}>
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
  <PopoverContent align="start" side="bottom" class="w-[280px] px-3.5 py-3">
    <div class="flex flex-col gap-y-1.5">
      <span class="text-xs font-medium truncate" title={item.title}>
        {item.title}
      </span>
      <MeasureFormattingControls id={item.id} {fmt} onChange={onFormatChange} />
    </div>
  </PopoverContent>
</Popover>
