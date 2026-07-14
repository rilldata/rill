<script lang="ts">
  import { tick } from "svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { createRuntimeServiceGitDiff } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import Diff2HtmlView from "@rilldata/web-common/components/diff/Diff2HtmlView.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { inferResourceKind } from "@rilldata/web-common/features/entity-management/infer-resource-kind";
  import { getIconComponent } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { get } from "svelte/store";

  // initialPath, when set, scrolls the diff to that file once the dialog has rendered.
  let {
    open = $bindable(false),
    remoteBranch,
    initialPath = undefined,
  }: {
    open?: boolean;
    remoteBranch: string | undefined;
    initialPath?: string | undefined;
  } = $props();

  const client = useRuntimeClient();
  // includeDiff fetches the combined patch alongside the file list. fetch is left false so this
  // reuses the fetch the changed-files list call already performed when the popover opened, avoiding
  // a redundant fetch. Gated on `open` and refetchOnMount "always" so it loads fresh each time the
  // dialog is opened.
  let diffQuery = $derived(
    createRuntimeServiceGitDiff(
      client,
      { remoteBranch, includeDiff: true },
      { query: { enabled: open && !!remoteBranch, refetchOnMount: "always" } },
    ),
  );
  let changedFiles = $derived($diffQuery.data?.changedFiles ?? []);
  let diff = $derived($diffQuery.data?.diff ?? "");
  let isFetching = $derived($diffQuery.isFetching);

  let diffPane = $state<HTMLElement | undefined>(undefined);

  // Map from subpath-relative file path → its .d2h-file-wrapper element.
  // Built once after the diff renders; rebuilt when diff or changedFiles changes.
  // diff2html names may carry the project subpath prefix, so we match by suffix
  // when building the map rather than on every scroll call.
  let wrapperMap = new Map<string, HTMLElement>();

  $effect(() => {
    const files = changedFiles;
    const pane = diffPane;
    const scrollPath = open ? initialPath : undefined;
    if (!diff || !pane) {
      wrapperMap = new Map();
      return;
    }
    void tick().then(() => {
      const map = new Map<string, HTMLElement>();
      for (const wrapper of pane.querySelectorAll<HTMLElement>(
        ".d2h-file-wrapper",
      )) {
        const name = wrapper
          .querySelector(".d2h-file-name")
          ?.textContent?.trim();
        if (!name) continue;
        const file = files.find(
          (f) => name === f.path || name.endsWith("/" + f.path),
        );
        if (file?.path) map.set(file.path, wrapper);
      }
      wrapperMap = map;
      if (scrollPath) map.get(scrollPath)?.scrollIntoView({ block: "start" });
    });
  });

  function scrollToFile(path: string | undefined) {
    if (!path) return;
    wrapperMap.get(path)?.scrollIntoView({ block: "start" });
  }

  function getFileKind(path: string): ResourceKind | null | undefined {
    const filePath = "/" + path;
    if (!fileArtifacts.hasFileArtifact(filePath)) {
      return inferResourceKind(filePath, "");
    }
    const artifact = fileArtifacts.getFileArtifact(filePath);
    return (get(artifact.resourceName)?.kind ??
      get(artifact.inferredResourceKind)) as ResourceKind | null | undefined;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content
    class="flex flex-col gap-0 p-0 w-[90vw] max-w-screen-xl h-[85vh] max-h-[85vh]"
  >
    <Dialog.Header class="px-4 py-3 border-b border-gray-200 text-left">
      <Dialog.Title class="text-sm font-semibold leading-none tracking-normal">
        {m.edit_changes_review_title()}
        {#if changedFiles.length > 0}
          <span class="font-normal text-fg-secondary"
            >· {m.edit_changes_file_count({ count: changedFiles.length })}</span
          >
        {/if}
      </Dialog.Title>
    </Dialog.Header>

    {#if isFetching}
      <div class="state-message">
        <DelayedSpinner isLoading={true} size="16px" />
        <span>{m.edit_changes_loading_diff()}</span>
      </div>
    {:else if changedFiles.length === 0}
      <div class="state-message">{m.edit_changes_none_to_show()}</div>
    {:else}
      <div class="flex flex-1 min-h-0">
        <ul class="file-nav">
          {#each changedFiles as file (file.path)}
            {@const IconComponent = getIconComponent(
              getFileKind(file.path ?? "") ?? undefined,
              "/" + (file.path ?? ""),
            )}
            <li>
              <button
                type="button"
                class="file-row"
                onclick={() => scrollToFile(file.path)}
              >
                <IconComponent />
                <span class="file-path" title={file.path}>{file.path}</span>
              </button>
            </li>
          {/each}
        </ul>
        <div class="diff-pane" bind:this={diffPane}>
          <Diff2HtmlView {diff} showFileHeaders>
            {#snippet empty()}
              <div class="state-message">{m.edit_changes_no_diff()}</div>
            {/snippet}
          </Diff2HtmlView>
        </div>
      </div>
    {/if}

    <Dialog.Footer class="px-4 py-3 border-t border-gray-200">
      <Button type="secondary" onClick={() => (open = false)}
        >{m.common_close()}</Button
      >
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .state-message {
    @apply flex flex-1 items-center justify-center gap-x-2;
    @apply text-xs text-fg-secondary;
  }

  .file-nav {
    @apply flex-none w-60 overflow-y-auto;
    @apply border-r border-gray-200 bg-surface-subtle;
    @apply flex flex-col gap-y-0.5 p-2;
  }

  .file-row {
    @apply flex items-center gap-x-2 w-full text-left;
    @apply px-2 py-1 rounded text-xs text-fg-secondary;
    @apply hover:bg-gray-100 hover:text-fg-primary;
  }

  .file-path {
    @apply truncate;
  }

  .diff-pane {
    @apply flex-1 overflow-auto;
  }
</style>
