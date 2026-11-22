<script lang="ts">
  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { escapeHtml } from "@rilldata/web-common/lib/i18n";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { DeleteTarget } from "./delete-target";

  // target is what the delete will remove. The host keeps it set while the dialog closes,
  // so the text does not change during the exit animation.
  // onConfirm runs the delete; the dialog closes itself first.
  let {
    open = $bindable(false),
    target,
    onConfirm,
  }: {
    open?: boolean;
    target: DeleteTarget | undefined;
    onConfirm: (target: DeleteTarget) => void;
  } = $props();

  // The file explorer runs in Rill Developer and in the Cloud editor, and only the latter can restore a deleted file.
  const { adminServer: isCloudEditor } = featureFlags;

  let titleKind = $derived(
    target?.type === "folder" ? "folder" : (target?.resource?.kind ?? "file"),
  );

  // The messages are static and the interpolated names are escaped, so the result is safe to render as HTML.
  let description = $derived.by(() => {
    if (!target) return "";
    const path = `<span class="path">${escapeHtml(target.path)}</span>`;
    if (target.type === "folder") {
      return m.file_explorer_delete_folder_description({ path });
    }
    if (!target.resource) {
      return m.file_explorer_delete_file_description({ path });
    }
    return m.file_explorer_delete_resource_description({
      kind: target.resource.kind,
      name: `<span class="name">${escapeHtml(target.resource.name)}</span>`,
      path,
    });
  });
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <!-- min-w-0 lets long paths wrap or truncate instead of widening the dialog's grid track -->
    <AlertDialogHeader class="min-w-0">
      <AlertDialogTitle>
        {m.file_explorer_delete_title({ kind: titleKind })}
      </AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-2">
        <div class="target">{@html description}</div>
        {#if target?.type === "folder" && target.files.length}
          <ul class="file-list">
            <!-- Shown relative to the folder, so truncation keeps the file names readable -->
            {#each target.files as file (file)}
              <li title={file}>{file.slice(target.path.length + 1)}</li>
            {/each}
          </ul>
        {/if}
        <div>
          {$isCloudEditor
            ? m.file_explorer_delete_undo_cloud()
            : m.file_explorer_delete_undo_developer()}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <AlertDialogCancel>
        {#snippet child({ props })}
          <Button {...props} large type="tertiary">{m.common_cancel()}</Button>
        {/snippet}
      </AlertDialogCancel>
      <AlertDialogAction>
        {#snippet child({ props })}
          <Button
            {...props}
            large
            type="destructive"
            onClick={() => {
              open = false;
              if (target) onConfirm(target);
            }}
          >
            {m.common_delete()}
          </Button>
        {/snippet}
      </AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>

<style lang="postcss">
  .target :global(.name) {
    @apply font-medium text-fg-primary;
  }

  .target :global(.path) {
    @apply font-mono text-fg-primary;
    /* A path has no spaces; let it wrap, but only when it is longer than a line. */
    overflow-wrap: anywhere;
  }

  .file-list {
    @apply flex flex-col gap-y-0.5;
    @apply max-h-40 overflow-y-auto;
    @apply text-xs text-fg-secondary font-mono;
  }

  .file-list li {
    /* shrink-0 keeps items at full height when the list overflows max-h; without it,
       truncate's overflow:hidden lets flex shrink the items so they overlap instead of scrolling */
    @apply truncate shrink-0;
  }
</style>
