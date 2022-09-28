<script lang="ts">
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import SlidingWords from "../tooltip/SlidingWords.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import { CATEGORICALS, NUMERICS, TIMESTAMPS } from "../../duckdb-data-types";

  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  export let name;
  export let type;
  export let totalRows;

  export let active = false;

  let titleTooltip;
</script>

<Tooltip
  location="right"
  alignment="center"
  distance={40}
  bind:active={titleTooltip}
>
  <!-- Wrap in a traditional div then force the ellipsis overflow in the child element.
                this will make the tooltip bound to the parent element while the child element can flow more freely
                and create the ellipisis due to the overflow.
            -->
  <div style:width="100%">
    <div
      class="column-profile-name text-ellipsis overflow-hidden whitespace-nowrap"
    >
      {name}
    </div>
  </div>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        {name}
      </svelte:fragment>
      <svelte:fragment slot="description">
        {type}
      </svelte:fragment>
    </TooltipTitle>

    {#if totalRows}
      <TooltipShortcutContainer>
        <SlidingWords {active}>
          {#if CATEGORICALS.has(type)}
            the top 10 values
          {:else if TIMESTAMPS.has(type)}
            row count over time
          {:else if NUMERICS.has(type)}
            the distribution of values
          {/if}
        </SlidingWords>
        <Shortcut>Click</Shortcut>

        <div>
          <StackingWord key="shift">copy</StackingWord>
          column name to clipboard
        </div>
        <Shortcut>
          <span style="font-family: var(--system);">⇧</span> + Click
        </Shortcut>
      </TooltipShortcutContainer>
    {:else}
      <!-- no data is available, so let's give a useful message-->
      no rows selected
    {/if}
  </TooltipContent>
</Tooltip>
