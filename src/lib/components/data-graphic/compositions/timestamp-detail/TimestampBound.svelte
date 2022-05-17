<script lang="ts">
  /**
   * TimestampBound.svelte
   * ---------------------
   * This component will render a bound on the TimestampDetail.svelte graph.
   * It also enables a shift + click to copy the bound as a query-ready timestamp.
   */
  import notifications from "$lib/components/notifications";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import { datePortion, timePortion } from "$lib/util/formatters";
  import { createShiftClickAction } from "$lib/util/shift-click-action";

  const { shiftClickAction } = createShiftClickAction();

  export let value: Date;
  export let label: string = "value";
  export let align: "left" | "right" = "left";
</script>

<Tooltip alignment={align == "left" ? "start" : "end"} distance={8}>
  <button
    class="text-{align} text-gray-500"
    style:line-height={1.1}
    use:shiftClickAction
    on:shift-click={async (event) => {
      let exportedValue = `TIMESTAMP '${value.toISOString()}'`;
      await navigator.clipboard.writeText(exportedValue);
      notifications.send({ message: `copied ${exportedValue} to clipboard` });
      // update this to set the active animation in the tooltip text
    }}
  >
    <div>
      {datePortion(value)}
    </div>
    <div>
      {timePortion(value)}
    </div>
  </button>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">{value.toISOString()}</svelte:fragment>
      <svelte:fragment slot="description">{label}</svelte:fragment>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        <StackingWord>copy</StackingWord> to clipboard
      </div>
      <Shortcut>
        <span
          style="
          font-family: var(--system); 
          font-size: 11.5px;
        ">⇧</span
        > + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
