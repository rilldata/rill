<script lang="ts">
  import { createRuntimeServiceRefreshAndReconcile } from "@rilldata/web-common/runtime-client";
  import { fade } from "svelte/transition";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { page } from "$app/stores";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import ModelWorkspaceCtAs from "../../models/workspace/ModelWorkspaceCTAs.svelte";
  import SourceCTAs from "./SourceCTAs.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();

  export let hasErrors: boolean;
  export let hasUnsavedChanges: boolean;
  export let assetName: string;
  export let resourceIsReconciling: boolean;
  export let refreshedOn: string | undefined;
  export let isLocalFileConnector: boolean;
  export let type: "source" | "model";

  $: context = $page.url.pathname;
  $: workspaceLayout = workspaces.get(context);
  $: tableVisible = workspaceLayout.table.visible;

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
</script>

<WorkspaceHeader titleInput={assetName} {hasUnsavedChanges} on:change>
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

    <IconButton on:click={workspaceLayout.table.toggle}>
      <span class="text-gray-500">
        <HideBottomPane size="18px" />
      </span>
      <svelte:fragment slot="tooltip-content">
        <SlidingWords active={$tableVisible} reverse>
          results preview
        </SlidingWords>
      </svelte:fragment>
    </IconButton>
  </svelte:fragment>

  <svelte:fragment slot="cta" let:width>
    {@const collapse = width < 800}

    <div class="flex gap-x-2 items-center">
      {#if type === "source"}
        <SourceCTAs
          {hasUnsavedChanges}
          {collapse}
          {hasErrors}
          {isLocalFileConnector}
          isReconciling={resourceIsReconciling}
          on:create-model
          on:refresh-source
          on:replace-source
          on:revert-source
          on:save-source
        />
      {:else}
        <ModelWorkspaceCtAs
          {collapse}
          modelHasError={hasErrors}
          modelName={assetName}
        />
      {/if}
    </div>
  </svelte:fragment>
</WorkspaceHeader>
