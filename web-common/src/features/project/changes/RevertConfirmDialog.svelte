<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // files are the subpath-relative paths that will be reverted. loading reflects the in-flight
  // revert so the confirm button shows a spinner and both buttons disable. onConfirm performs the
  // revert; the host closes the dialog by flipping `open` once it resolves.
  let {
    open = $bindable(false),
    files,
    loading = false,
    onConfirm,
  }: {
    open?: boolean;
    files: string[];
    loading?: boolean;
    onConfirm: () => void;
  } = $props();
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle
        >{m.edit_changes_revert_confirm_title()}</AlertDialogTitle
      >
      <AlertDialogDescription>
        <div class="mt-1">
          {m.edit_changes_revert_confirm_body({ count: files.length })}
        </div>
        <ul class="file-list">
          {#each files as file (file)}
            <li title={file}>{file}</li>
          {/each}
        </ul>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" disabled={loading} onClick={() => (open = false)}
        >{m.common_cancel()}</Button
      >
      <Button
        type="destructive"
        disabled={loading}
        {loading}
        onClick={onConfirm}>{m.edit_changes_revert()}</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>

<style lang="postcss">
  .file-list {
    @apply mt-2 flex flex-col gap-y-0.5;
    @apply max-h-40 overflow-y-auto;
    @apply text-xs text-fg-secondary font-mono;
  }

  .file-list li {
    /* shrink-0 keeps items at full height when the list overflows max-h; without it,
       truncate's overflow:hidden lets flex shrink the items so they overlap instead of scrolling */
    @apply truncate shrink-0;
  }
</style>
