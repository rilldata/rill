<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { createEventDispatcher } from "svelte";
  import EnterIcon from "../../../components/icons/EnterIcon.svelte";
  import ButtonContent from "./ButtonContent.svelte";

  const dispatch = createEventDispatcher();

  export let hasErrors: boolean;
  export let hasUnsavedChanges: boolean;
  export let isLocalFileConnector: boolean;
  export let collapse: boolean;

  $: label = hasUnsavedChanges ? "Save and refresh" : "Refresh";

  $: type = (
    hasUnsavedChanges ? "primary" : "secondary"
  ) as Button["$$prop_def"]["type"];
</script>

<Button
  type="secondary"
  disabled={!hasUnsavedChanges}
  on:click={() => dispatch("revert-source")}
>
  <IconSpaceFixer pullLeft pullRight={collapse}>
    <UndoIcon size="14px" />
  </IconSpaceFixer>
  <ResponsiveButtonText {collapse}>Revert changes</ResponsiveButtonText>
</Button>

{#if !isLocalFileConnector || hasUnsavedChanges}
  <Button
    on:click={() => {
      if (isLocalFileConnector && !hasUnsavedChanges) return;
      if (hasUnsavedChanges) {
        dispatch("save-source");
      } else {
        dispatch("refresh-source");
      }
    }}
    {label}
    {type}
  >
    <ButtonContent {collapse} {hasUnsavedChanges} {isLocalFileConnector} />
  </Button>
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button builders={[builder]} {label} {type}>
        <ButtonContent {collapse} {hasUnsavedChanges} {isLocalFileConnector} />
      </Button>
    </DropdownMenu.Trigger>

    <DropdownMenu.Content>
      <DropdownMenu.Item
        on:click={() => {
          dispatch("refresh-source");
        }}
      >
        Refresh source
      </DropdownMenu.Item>
      <DropdownMenu.Item on:click={() => dispatch("replace-source")}>
        Replace source with uploaded file
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<Button
  disabled={hasUnsavedChanges || hasErrors}
  on:click={() => dispatch("create-model")}
  type="brand"
>
  <ResponsiveButtonText {collapse}>Create model</ResponsiveButtonText>
  <IconSpaceFixer pullLeft pullRight={collapse}>
    <EnterIcon size="14px" />
  </IconSpaceFixer>
</Button>
