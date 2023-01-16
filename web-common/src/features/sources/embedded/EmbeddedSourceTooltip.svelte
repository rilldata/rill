<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let sourceName: string;
  export let connector = undefined;
</script>

<TooltipTitle>
  <div slot="name" class="break-all">
    {sourceName}
  </div>
  <svelte:fragment slot="description">
    {#if connector}
      {connector}
    {/if}
  </svelte:fragment>
</TooltipTitle>
<TooltipShortcutContainer>
  <div>
    <StackingWord key="shift">Copy</StackingWord> name to clipboard
  </div>
  <Shortcut>
    <span
      style="
      font-family: var(--system); 
      font-size: 11.5px;
    ">⇧</span
    > + Click</Shortcut
  >
</TooltipShortcutContainer>
