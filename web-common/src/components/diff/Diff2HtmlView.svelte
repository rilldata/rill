<!--
  Renders a git patch as HTML using diff2html, with Rill's diff styling.
  Handles a single-file patch (chat) or a combined multi-file patch (review dialog).
-->
<script lang="ts">
  import { html as diffToHtml } from "diff2html";
  import "diff2html/bundles/css/diff2html.min.css";
  import DOMPurify from "dompurify";

  export let diff: string;
  // showFileHeaders keeps diff2html's per-file header visible (and sticky) so each
  // section is labeled. Callers that render their own header leave it false.
  export let showFileHeaders = false;
  // scrollX makes this element the horizontal scroll container. Leave false when an
  // ancestor already handles scrolling (else the computed overflow-y would break
  // sticky file headers).
  export let scrollX = false;

  $: diffHtml = diff
    ? DOMPurify.sanitize(
        diffToHtml(diff, {
          drawFileList: false,
          outputFormat: "line-by-line",
          matching: "lines",
        }),
      )
    : "";
</script>

{#if diffHtml}
  <div
    class="diff-view"
    class:with-file-headers={showFileHeaders}
    class:scroll-x={scrollX}
  >
    {@html diffHtml}
  </div>
{:else}
  <slot name="empty" />
{/if}

<style lang="postcss">
  .diff-view.scroll-x {
    @apply overflow-x-auto;
  }

  .diff-view :global(.d2h-file-wrapper) {
    border: none;
    border-radius: 0;
    margin: 0;
  }

  /* File header: hidden by default (callers draw their own), sticky when shown. */
  .diff-view:not(.with-file-headers) :global(.d2h-file-header) {
    display: none;
  }

  .diff-view.with-file-headers :global(.d2h-file-header) {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 6px 12px;
    background-color: var(--color-gray-100);
    border-bottom: 1px solid var(--color-gray-200);
  }

  .diff-view.with-file-headers :global(.d2h-file-name) {
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    color: var(--color-gray-900);
  }

  .diff-view :global(.d2h-wrapper) {
    font-size: 12px;
    line-height: 20px;
  }

  .diff-view :global(.d2h-diff-table) {
    width: max-content;
    min-width: 100%;
    border-collapse: collapse;
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas,
      "Liberation Mono", monospace;
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

  /* Addition lines */
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

  /* Deletion lines */
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

  /* Context/unchanged lines */
  .diff-view :global(.d2h-cntx) {
    background-color: var(--surface-subtle);
  }

  .diff-view :global(.d2h-cntx .d2h-code-linenumber) {
    background-color: var(--color-gray-100);
  }

  .diff-view :global(.d2h-cntx .d2h-code-line-prefix),
  .diff-view :global(.d2h-cntx .d2h-code-line-ctn) {
    background-color: var(--surface-subtle);
  }

  /* Hunk header (@@ … @@) */
  .diff-view :global(.d2h-info) {
    background-color: var(--color-blue-100);
    color: var(--color-gray-600);
    padding: 6px 10px;
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    line-height: 20px;
  }
</style>
