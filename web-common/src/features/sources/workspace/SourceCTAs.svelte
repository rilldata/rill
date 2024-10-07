<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { createEventDispatcher } from "svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";

  const dispatch = createEventDispatcher();

  export let hasErrors: boolean;
  export let hasUnsavedChanges: boolean;
  export let isLocalFileConnector: boolean;

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
      label="Refresh"
      type="secondary"
      disabled={hasUnsavedChanges}
    >
      <RefreshIcon size="14px" />
    </Button>

    <TooltipContent slot="tooltip-content">
      {#if hasUnsavedChanges}
        Save your changes to refresh
      {:else}
        Refresh source
      {/if}
    </TooltipContent>
  </Tooltip>
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Tooltip distance={8}>
        <Button builders={[builder]} label="Refresh" {type}>
          <RefreshIcon size="14px" />
        </Button>
        <TooltipContent slot="tooltip-content">Refresh source</TooltipContent>
      </Tooltip>
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
