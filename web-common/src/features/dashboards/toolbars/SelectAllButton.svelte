<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { Inspect } from "lucide-svelte";
  import { Button } from "../../../components/button";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let disabled = false;
  export let areAllTableRowsSelected: boolean;
  export let onToggleSearchItems: () => void;
</script>

<Tooltip distance={4} location="top">
  <TooltipContent slot="tooltip-content">
    {#if areAllTableRowsSelected}
      <div>{m.dashboard_deselect_all_selections()}</div>
    {:else}
      <TooltipShortcutContainer pad={false}>
        <div>{m.dashboard_select_all()}</div>
        <Shortcut
          ><span style="font-family: var(--system);"> ⌘ </span> + A</Shortcut
        >
      </TooltipShortcutContainer>
    {/if}
  </TooltipContent>
  <Button type="toolbar" onClick={onToggleSearchItems} {disabled}>
    <div class="text-fg-secondary">
      <Inspect size={16} />
    </div>
    {areAllTableRowsSelected && !disabled
      ? m.dashboard_deselect_all()
      : m.dashboard_select_all()}
  </Button>
</Tooltip>
