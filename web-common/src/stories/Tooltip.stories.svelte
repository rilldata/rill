<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  // import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
</script>

<Meta title="tooltip template" />

<Template let:args>
  <div class="px-5 mx-3 border-lime-500 stroke-orange-800">styled content</div>

  <div class="px-5 mx-3 ui-copy-strong">styled content</div>
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
</Template>

<Story name="Rounded" args={{ rounded: true }} />

<Story name="Square" source args={{ rounded: false }} />

<!-- Dynamic snippet should be disabled for this story -->
