<script lang="ts">
  // import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let Hst;
</script>

<Hst.Story>
  <Hst.Variant>
    faksjdfl;aj
    <TooltipContent maxWidth="360px">
      inside tooltip
      <TooltipTitle>
        foo asdjflj
        <!-- <FormattedDataType
          slot="name"
          value={12.23}
          type={"RILL_PERCENTAGE_CHANGE"}
          dark
        /> -->
      </TooltipTitle>
      <TooltipShortcutContainer>
        <div>
          <StackingWord key="shift">Copy</StackingWord> this value to clipboard
        </div>
        <Shortcut>
          <span style="font-family: var(--system);">⇧</span> + Click
        </Shortcut>
      </TooltipShortcutContainer>
    </TooltipContent>
  </Hst.Variant>
</Hst.Story>
