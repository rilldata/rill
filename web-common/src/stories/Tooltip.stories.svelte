<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
</script>

<Meta title="tooltip template" />

<Template let:args>
  <TooltipContent maxWidth="360px">
    <TooltipTitle>
      <div slot="name">{args.name}</div>
      <div slot="description">{args.description}</div>
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

<Story name="base case" args={{ name: "name", description:"description" }} />

<Story name="long name" args={{ name: "A_very_long_string_in_the_name_field_that_has_no_whitespace_like_a_terrible_database_column_name" ,description:"description"}} />

<Story name="long description" source args={{ name:"name",description: "A long string in the description field that does have whitespace like a longish string in a database field" }} />

