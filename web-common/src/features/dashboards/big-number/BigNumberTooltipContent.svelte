<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipDescription from "@rilldata/web-common/components/tooltip/TooltipDescription.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";

  export let measure: MetricsViewSpecMeasure;
  export let value = "";

  $: description =
    measure?.description || measure?.displayName || measure?.expression;
  $: name = measure?.displayName || measure?.expression;
</script>

<TooltipContent maxWidth="280px">
  <TooltipTitle>
    <svelte:fragment slot="name">
      {name}
    </svelte:fragment>
    <svelte:fragment slot="description">
      {value}
    </svelte:fragment>
  </TooltipTitle>

  <TooltipDescription>
    {description}
  </TooltipDescription>

  <TooltipShortcutContainer>
    <div>
      <StackingWord key="shift">Copy</StackingWord> to clipboard
    </div>
    <Shortcut>
      <span style="font-family: var(--system);">⇧</span> + Click
    </Shortcut>
  </TooltipShortcutContainer>
</TooltipContent>
