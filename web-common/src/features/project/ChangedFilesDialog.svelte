<script lang="ts">
  import { tick } from "svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { createRuntimeServiceGitDiff } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { html as diffToHtml } from "diff2html";
  import "diff2html/bundles/css/diff2html.min.css";
  import DOMPurify from "dompurify";
  import FileChangeBadge from "./FileChangeBadge.svelte";

  export let open = false;
  export let remoteBranch: string | undefined;
  // initialPath, when set, scrolls the diff to that file once the dialog has rendered.
  export let initialPath: string | undefined = undefined;

  const client = useRuntimeClient();
  // includeDiff fetches the combined patch alongside the file list. fetch is left false so this
  // reuses the fetch the changed-files list call already performed when the popover opened, avoiding
  // a redundant fetch. Gated on `open` and refetchOnMount "always" so it loads fresh each time the
  // dialog is opened.
  $: diffQuery = createRuntimeServiceGitDiff(
    client,
    { remoteBranch, includeDiff: true },
    { query: { enabled: open && !!remoteBranch, refetchOnMount: "always" } },
  );
  $: changedFiles = $diffQuery.data?.changedFiles ?? [];
  $: diff = $diffQuery.data?.diff ?? "";
  $: isFetching = $diffQuery.isFetching;

  // Reuse the same diff2html config the AI chat diff view uses.
  $: diffHtml = diff
    ? DOMPurify.sanitize(
        diffToHtml(diff, {
          drawFileList: false,
          outputFormat: "line-by-line",
          matching: "lines",
        }),
      )
    : "";

  let diffPane: HTMLElement | undefined;

  // Scroll the diff to a file's section. diff2html's rendered file name may carry the project
  // subpath prefix, so match it by suffix against the subpath-relative changed-files path.
  // Scroll the diff pane directly (rather than scrollIntoView, which also scrolls ancestors and
  // lands imprecisely against the sticky file headers) so the file's header sits at the top.
  function scrollToFile(path: string | undefined) {
    if (!path || !diffPane) return;
    const wrappers =
      diffPane.querySelectorAll<HTMLElement>(".d2h-file-wrapper");
    for (const wrapper of wrappers) {
      const name = wrapper.querySelector(".d2h-file-name")?.textContent?.trim();
      if (name === path || name?.endsWith("/" + path)) {
        const top =
          wrapper.getBoundingClientRect().top -
          diffPane.getBoundingClientRect().top +
          diffPane.scrollTop;
        diffPane.scrollTo({ top });
        return;
      }
    }
  }

  // Jump to the file the user clicked once the diff has rendered.
  $: if (open && diffHtml && initialPath) {
    void tick().then(() => scrollToFile(initialPath));
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content
    class="flex flex-col gap-0 p-0 w-[90vw] max-w-screen-xl h-[85vh] max-h-[85vh]"
  >
    <Dialog.Header class="px-4 py-3 border-b border-gray-200 text-left">
      <Dialog.Title class="text-sm font-semibold leading-none tracking-normal">
        Review changes
        {#if changedFiles.length > 0}
          <span class="font-normal text-fg-secondary"
            >· {changedFiles.length}
            {changedFiles.length === 1 ? "file" : "files"}</span
          >
        {/if}
      </Dialog.Title>
    </Dialog.Header>

    {#if isFetching}
      <div class="state-message">
        <DelayedSpinner isLoading={true} size="16px" />
        <span>Loading diff…</span>
      </div>
    {:else if changedFiles.length === 0}
      <div class="state-message">No changes to show</div>
    {:else}
      <div class="flex flex-1 min-h-0">
        <ul class="file-nav">
          {#each changedFiles as file (file.path)}
            <li>
              <button
                type="button"
                class="file-row"
                onclick={() => scrollToFile(file.path)}
              >
                <FileChangeBadge status={file.status} />
                <span class="file-path" title={file.path}>{file.path}</span>
              </button>
            </li>
          {/each}
        </ul>
        <div class="diff-pane" bind:this={diffPane}>
          {#if diffHtml}
            <div class="diff-view">
              {@html diffHtml}
            </div>
          {:else}
            <div class="state-message">No diff to display</div>
          {/if}
        </div>
      </div>
    {/if}

    <Dialog.Footer class="px-4 py-3 border-t border-gray-200">
      <Button type="secondary" onClick={() => (open = false)}>Close</Button>
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

  /* diff2html styling, adapted from the chat FileDiffBlock; file headers stay visible so each
     file's section is labeled in the combined diff. */
  .diff-view :global(.d2h-file-wrapper) {
    border: none;
    border-radius: 0;
    margin: 0;
  }

  .diff-view :global(.d2h-file-header) {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 6px 12px;
    background-color: var(--color-gray-100);
    border-bottom: 1px solid var(--color-gray-200);
  }

  .diff-view :global(.d2h-file-name) {
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    color: var(--color-gray-900);
  }

  .diff-view :global(.d2h-diff-table) {
    width: max-content;
    min-width: 100%;
    border-collapse: collapse;
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
  }

  .diff-view :global(.d2h-diff-tbody tr) {
    border: none;
    line-height: 20px;
  }

  .diff-view :global(.d2h-code-line) {
    padding: 0;
    border: none;
  }

  /* Show only the new line-number column. */
  .diff-view :global(td.d2h-code-linenumber:first-child) {
    display: none;
  }

  .diff-view :global(.d2h-code-linenumber) {
    position: static;
    color: var(--color-gray-500);
    text-align: right;
    padding: 0 10px;
    min-width: 40px;
    width: 40px;
    background-color: var(--color-gray-100);
    user-select: none;
    vertical-align: top;
    white-space: nowrap;
    border-right: 1px solid var(--color-gray-300);
  }

  .diff-view :global(.d2h-code-line-prefix) {
    position: static;
    padding: 0 8px;
    user-select: none;
    width: 20px;
    min-width: 20px;
    text-align: center;
    vertical-align: top;
    color: var(--color-gray-500);
  }

  .diff-view :global(.d2h-code-line-ctn) {
    position: static;
    padding: 0 8px 0 0;
    white-space: pre;
    word-wrap: normal;
    vertical-align: top;
    color: var(--color-gray-900);
  }

  .diff-view :global(.d2h-ins) {
    background-color: var(--color-green-100);
  }

  .diff-view :global(.d2h-ins .d2h-code-linenumber) {
    background-color: var(--color-green-200);
  }

  .diff-view :global(.d2h-ins .d2h-code-line-prefix) {
    color: var(--color-green-700);
    background-color: var(--color-green-100);
  }

  .diff-view :global(.d2h-ins .d2h-code-line-ctn) {
    background-color: var(--color-green-100);
  }

  .diff-view :global(.d2h-ins ins) {
    background-color: var(--color-green-400);
  }

  .diff-view :global(.d2h-del) {
    background-color: var(--color-red-100);
  }

  .diff-view :global(.d2h-del .d2h-code-linenumber) {
    background-color: var(--color-red-200);
  }

  .diff-view :global(.d2h-del .d2h-code-line-prefix) {
    color: var(--color-red-600);
    background-color: var(--color-red-100);
  }

  .diff-view :global(.d2h-del .d2h-code-line-ctn) {
    background-color: var(--color-red-100);
  }

  .diff-view :global(.d2h-del del) {
    background-color: var(--color-red-400);
  }

  .diff-view :global(.d2h-cntx .d2h-code-linenumber) {
    background-color: var(--color-gray-100);
  }

  /* Hunk header (@@ … @@) */
  .diff-view :global(.d2h-info) {
    background-color: var(--color-blue-100);
    color: var(--color-gray-600);
    padding: 4px 10px;
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    line-height: 20px;
  }
</style>
