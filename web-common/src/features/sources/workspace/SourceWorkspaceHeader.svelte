<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import EnterIcon from "../../../components/icons/EnterIcon.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import { createRuntimeServiceRefreshAndReconcile } from "@rilldata/web-common/runtime-client";
  import { fade, slide } from "svelte/transition";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { createEventDispatcher } from "svelte";
  import { quintOut } from "svelte/easing";

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const dispatch = createEventDispatcher();

  export let hasErrors: boolean;
  export let isSourceUnsaved: boolean;
  export let sourceName: string;
  export let sourceIsReconciling: boolean;
  export let refreshedOn: string | undefined;
  export let isLocalFileConnector: boolean;

  function formatRefreshedOn(refreshedOn: string) {
    const date = new Date(refreshedOn);
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<WorkspaceHeader titleInput={sourceName} {isSourceUnsaved} on:change>
  <svelte:fragment slot="workspace-controls">
    {#if $refreshSourceMutation.isLoading}
      Refreshing...
    {:else}
      <div class="flex items-center pr-2 gap-x-2">
        {#if refreshedOn}
          <div
            class="ml-2 ui-copy-muted line-clamp-2"
            style:font-size="11px"
            transition:fade={{ duration: 200 }}
          >
            Ingested on {formatRefreshedOn(refreshedOn)}
          </div>
        {/if}
      </div>
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="cta" let:width>
    {@const collapse = isHeaderWidthSmall(width)}
    <PanelCTA side="right">
      <Button
        type="secondary"
        disabled={!isSourceUnsaved}
        on:click={() => dispatch("revert")}
      >
        <IconSpaceFixer pullLeft pullRight={collapse}>
          <UndoIcon size="14px" />
        </IconSpaceFixer>
        <ResponsiveButtonText {collapse}>Revert changes</ResponsiveButtonText>
      </Button>

      <DropdownMenu.Root>
        <DropdownMenu.Trigger
          disabled={!isLocalFileConnector || isSourceUnsaved}
        >
          <Button
            div={isLocalFileConnector}
            role={isLocalFileConnector ? "presentation" : "button"}
            on:click={() => {
              if (isLocalFileConnector && !isSourceUnsaved) return;

              if (isSourceUnsaved) {
                dispatch("save");
              } else {
                dispatch("refresh-source");
              }
            }}
            disabled={sourceIsReconciling}
            label={isSourceUnsaved ? "Save and refresh" : "Refresh"}
            type={isSourceUnsaved ? "primary" : "secondary"}
          >
            <IconSpaceFixer pullLeft pullRight={collapse}>
              <RefreshIcon size="14px" />
            </IconSpaceFixer>
            <ResponsiveButtonText {collapse}>
              <div class="flex">
                {#if isSourceUnsaved}
                  <span
                    class="pr-1 w-fit whitespace-nowrap"
                    transition:slide={{
                      duration: 250,
                      axis: "x",
                      easing: quintOut,
                    }}
                  >
                    Save and
                  </span>
                {/if}
                <span class:lowercase={isSourceUnsaved}>Refresh</span>
              </div>
            </ResponsiveButtonText>
            {#if !isSourceUnsaved && isLocalFileConnector}
              <CaretDownIcon size="14px" />
            {/if}
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

      <Button
        disabled={isSourceUnsaved || hasErrors}
        on:click={() => dispatch("create-model")}
      >
        <ResponsiveButtonText {collapse}>Create model</ResponsiveButtonText>
        <IconSpaceFixer pullLeft pullRight={collapse}>
          <EnterIcon size="14px" />
        </IconSpaceFixer>
      </Button>
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>
