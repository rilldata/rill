<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Inspect } from "lucide-svelte";
  import { Button } from "../../../components/button";

  export let disabled = false;
  export let areAllTableRowsSelected: boolean;
  export let onToggleSearchItems: () => void;
</script>

<Tooltip distance={4} location="top">
  <TooltipContent slot="tooltip-content">
    {#if areAllTableRowsSelected}
      <div>{m.dashboards_toolbar_deselect_all_tooltip()}</div>
    {:else}
      <TooltipShortcutContainer pad={false}>
        <div>{m.dashboards_toolbar_select_all()}</div>
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
      ? m.dashboards_toolbar_deselect_all()
      : m.dashboards_toolbar_select_all()}
  </Button>
</Tooltip>
