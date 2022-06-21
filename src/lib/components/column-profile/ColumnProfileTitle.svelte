<script lang="ts">
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import SlidingWords from "$lib/components/tooltip/SlidingWords.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import { CATEGORICALS, NUMERICS, TIMESTAMPS } from "$lib/duckdb-data-types";

  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
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
        <SlidingWords {active} hovered={titleTooltip}>
          {#if CATEGORICALS.has(type)}
            the top 10 values
          {:else if TIMESTAMPS.has(type)}
            the count(*) over time
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
