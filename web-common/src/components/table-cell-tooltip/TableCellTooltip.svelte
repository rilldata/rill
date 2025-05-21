<script lang="ts">
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let label: string | number;
  export let selected: boolean;
  export let excluded: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let showInspect = true;
  export let showExclusiveSelect = true;
  export let cellLabel = "Filter value";
</script>

<TooltipContent>
  <TooltipTitle>
    <svelte:fragment slot="name">
      {label}
    </svelte:fragment>
  </TooltipTitle>

  <TooltipShortcutContainer>
    {#if atLeastOneActive}
      <div>
        {excluded ? "Include" : "Exclude"}
        {cellLabel}
      </div>
    {:else}
      <div class="text-ellipsis overflow-hidden whitespace-nowrap">
        Filter {filterExcludeMode ? "out" : "on"}
        {cellLabel}
      </div>
    {/if}
    <Shortcut>Click</Shortcut>
  </TooltipShortcutContainer>

  {#if isClipboardApiSupported()}
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord>
        value to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click on cell
      </Shortcut>
    </TooltipShortcutContainer>
  {/if}

  {#if showInspect}
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Inspect</StackingWord>
        value
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + I
      </Shortcut>
    </TooltipShortcutContainer>
  {/if}

  {#if showExclusiveSelect && !selected && atLeastOneActive}
    <TooltipShortcutContainer>
      <div>Exclusively select this value</div>
      <Shortcut>
        <MetaKey />
      </Shortcut>
    </TooltipShortcutContainer>
  {/if}
</TooltipContent>
