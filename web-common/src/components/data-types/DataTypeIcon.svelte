<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    copyToClipboard,
    isClipboardApiSupported,
  } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { fieldTypeToSymbol } from "@rilldata/web-common/lib/duckdb-data-types";
  import ShiftKey from "../tooltip/ShiftKey.svelte";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";

  export let color = "text-fg-secondary";
  export let type: string;
  export let suppressTooltip = false;
</script>

<Tooltip
  location="left"
  distance={16}
  suppress={suppressTooltip || !isClipboardApiSupported()}
>
  <button
    title={type}
    class="
    {color}
    grid place-items-center rounded"
    style="width: 16px; height: 16px;"
    on:click={modified({
      shift: () => copyToClipboard(type),
    })}
  >
    <div>
      <svelte:component this={fieldTypeToSymbol(type)} size="16px" />
    </div>
  </button>
  <TooltipContent maxWidth="300px" slot="tooltip-content">
    <TooltipTitle>
      <div slot="name" class="truncate">
        {type}
      </div>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord> type to clipboard
      </div>
      <Shortcut>
        <ShiftKey /> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
