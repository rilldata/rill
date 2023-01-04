<script lang="ts">
  import { goto } from "$app/navigation";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let sourceName;
  export let connector;
  export let embeds;

  function handleKeyDown(event: KeyboardEvent) {
    const numbers = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"];
    if (event.ctrlKey && numbers.includes(event.key)) {
      let number = +event.key - 1;
      if (number === -1) number = 9;
      if (embeds.length - 1 >= number) goto(`/model/${embeds[number]}`);
    }
  }
</script>

<svelte:window on:keydown={handleKeyDown} />

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
  <!-- display all the embed sources-->
  {#each embeds.slice(0, 9) as embeddedIn, i}
    <div>go to {embeddedIn} reference</div>
    <Shortcut>
      Ctrl + {i + 1}
    </Shortcut>
  {/each}
</TooltipShortcutContainer>
