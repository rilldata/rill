<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { createEventDispatcher } from "svelte";
  import EnterIcon from "../../../components/icons/EnterIcon.svelte";
  import ButtonContent from "./ButtonContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

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

{#if !isLocalFileConnector || hasUnsavedChanges}
  <Tooltip distance={8}>
    <Button
      square
      on:click={() => {
        if (isLocalFileConnector && !hasUnsavedChanges) return;
        if (hasUnsavedChanges) {
          dispatch("save-source");
        } else {
          dispatch("refresh-source");
        }
      }}
      {label}
      type="secondary"
      disabled={hasUnsavedChanges}
    >
      <ButtonContent {collapse} {hasUnsavedChanges} {isLocalFileConnector} />
    </Button>

    <TooltipContent slot="tooltip-content"
      >Save your changes to refresh
    </TooltipContent>
  </Tooltip>
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
  type="secondary"
>
  Create model
</Button>
