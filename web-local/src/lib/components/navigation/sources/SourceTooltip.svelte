<script>
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { isMac } from "@rilldata/web-local/lib/util/os-detection";

  export let sourceName;
</script>

<TooltipTitle>
  <svelte:fragment slot="name">
    {sourceName}
  </svelte:fragment>
  <svelte:fragment slot="description" />
</TooltipTitle>
<TooltipShortcutContainer>
  <div>
    <StackingWord key="command">Query</StackingWord> in workspace
  </div>
  <Shortcut>
    {#if isMac()}<span
        style="
      font-family: var(--system); 
      font-size: 11.5px;
    ">⌘</span
      >{:else}ctrl{/if} + Click</Shortcut
  >
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
