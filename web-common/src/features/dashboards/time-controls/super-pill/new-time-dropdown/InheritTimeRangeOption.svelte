<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { INHERIT_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";

  let {
    inherited,
    onSelect,
  }: {
    inherited: boolean;
    onSelect: (range: string) => void;
  } = $props();
</script>

<div class="w-full h-fit px-1">
  <button
    type="button"
    role="menuitem"
    class="group truncate h-7 p-2 text-popover-foreground justify-between overflow-hidden hover:bg-popover-accent rounded-sm w-full select-none flex items-center"
    onclick={() => onSelect(INHERIT_TIME_RANGE_ALIAS)}
  >
    <span class:font-bold={inherited}>{m.canvas_inherit_from_canvas()}</span>
  </button>
  <div class="h-px w-full bg-border my-1"></div>
</div>
