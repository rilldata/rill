<script lang="ts">
  import Shortcut from "../../tooltip/Shortcut.svelte";
  import StackingWord from "../../tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "../../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../../tooltip/TooltipTitle.svelte";

  export let modelName;
</script>

<TooltipTitle>
  <svelte:fragment slot="name">
    {modelName}
  </svelte:fragment>
  <svelte:fragment slot="description" />
</TooltipTitle>
<TooltipShortcutContainer>
  <div>open in workspace</div>
  <Shortcut>Click</Shortcut>
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
