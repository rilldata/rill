<script>
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { isClipboardApiSupported } from "../../../lib/actions/shift-click-action";

  function isMac() {
    return window.navigator.userAgent.includes("Macintosh");
  }
  export let sourceName;
  export let connector;
</script>

<TooltipTitle>
  <div slot="name" class="break-all">
    {sourceName}
  </div>
  <svelte:fragment slot="description">
    {connector}
  </svelte:fragment>
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
  {#if isClipboardApiSupported()}
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
  {/if}
</TooltipShortcutContainer>
