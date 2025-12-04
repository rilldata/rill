<!--
  Renders a file diff block with collapsible header and unified diff display.
  Used to show AI-made file changes in the chat interface.
  Styled to match other tool call blocks in the chat.
-->
<script lang="ts">
  import { html } from "diff2html";
  import "diff2html/bundles/css/diff2html.min.css";
  import DOMPurify from "dompurify";
  import CaretDownIcon from "../../../../components/icons/CaretDownIcon.svelte";
  import ChevronRight from "../../../../components/icons/ChevronRight.svelte";

  export let filePath: string;
  export let diff: string = "";
  export let isNewFile: boolean = false;

  // Start collapsed like other tool calls
  let isExpanded = false;

  // Generate and sanitize diff HTML
  $: diffHtml = diff
    ? DOMPurify.sanitize(
        html(diff, {
          drawFileList: false,
          outputFormat: "line-by-line",
          matching: "lines",
        }),
      )
    : "";

  function toggleExpanded() {
    isExpanded = !isExpanded;
  }

  function handleLinkClick(event: MouseEvent) {
    // Stop propagation so the header button doesn't also toggle
    event.stopPropagation();
  }

  // Get display name for the header
  $: fileName = filePath.split("/").pop() || filePath;
</script>

<div class="tool-container">
  <button class="tool-header" on:click={toggleExpanded}>
    <div class="tool-icon">
      {#if isExpanded}
        <CaretDownIcon size="16" />
      {:else}
        <ChevronRight size="16" />
      {/if}
    </div>
    <div class="tool-name">
      {isNewFile ? "Created" : "Modified"}: {fileName}
      {#if isNewFile}
        <span class="new-badge">New</span>
      {/if}
    </div>
  </button>

  {#if isExpanded}
    <div class="tool-content">
      <div class="tool-section">
        <div class="tool-section-title">
          <a
            href="/files{filePath}"
            class="file-path-link"
            on:click={handleLinkClick}
          >
            {filePath}
          </a>
        </div>
        <div class="tool-section-content">
          {#if diffHtml}
            <div class="diff-view">
              {@html diffHtml}
            </div>
          {:else}
            <div class="no-changes-message">No changes detected</div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  /* Match the tool-container styling from CallMessage.svelte */
  .tool-container {
    @apply w-full max-w-[90%] self-start;
    @apply border border-gray-200 rounded-lg bg-gray-50;
  }

  .tool-header {
    @apply w-full flex items-center gap-2 p-2;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-sm transition-colors;
  }

  .tool-header:hover {
    @apply bg-gray-100;
  }

  .tool-icon {
    @apply text-gray-500 flex items-center;
  }

  .tool-name {
    @apply font-medium text-gray-700 flex-1 text-left flex items-center gap-2;
  }

  .new-badge {
    @apply text-[0.625rem] px-1.5 py-0.5 rounded bg-green-100 text-green-700 font-medium uppercase;
  }

  .tool-content {
    @apply border-t border-gray-200 bg-white rounded-b-lg overflow-hidden;
  }

  .tool-section {
    @apply p-2;
  }

  .tool-section-title {
    @apply text-[0.625rem] font-semibold text-gray-500 mb-1.5;
    @apply uppercase tracking-wide;
  }

  .file-path-link {
    @apply text-primary-600 cursor-pointer;
    @apply normal-case tracking-normal font-mono;
  }

  .file-path-link:hover {
    @apply text-primary-800 underline;
  }

  .tool-section-content {
    @apply border border-gray-200 rounded-md overflow-hidden;
    background-color: #f6f8fa;
  }

  .no-changes-message {
    @apply p-3 text-xs text-gray-500 italic;
  }

  .diff-view {
    @apply overflow-x-auto;

    /*
     * GitHub-style diff theming
     * Note: diff2html CSS variables are defined at :root level and don't inherit
     * from parent elements, so we must use direct style overrides for scoped theming.
     */
  }

  /* Structural overrides */
  .diff-view :global(.d2h-file-header) {
    display: none;
  }

  .diff-view :global(.d2h-file-wrapper) {
    border: none;
    border-radius: 0;
    margin: 0;
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

  /* Hide the first (old) line number column - show only new line numbers */
  .diff-view :global(td.d2h-code-linenumber:first-child) {
    display: none;
  }

  .diff-view :global(.d2h-code-linenumber) {
    position: static;
    color: rgba(31, 35, 40, 0.5);
    text-align: right;
    padding: 0 10px;
    min-width: 40px;
    width: 40px;
    background-color: #f6f8fa;
    user-select: none;
    vertical-align: top;
    white-space: nowrap;
    border-right: 1px solid #d1d9e0;
  }

  .diff-view :global(.d2h-code-line-prefix) {
    position: static;
    padding: 0 8px;
    user-select: none;
    width: 20px;
    min-width: 20px;
    text-align: center;
    vertical-align: top;
    color: rgba(31, 35, 40, 0.5);
  }

  .diff-view :global(.d2h-code-line-ctn) {
    position: static;
    padding: 0 8px 0 0;
    white-space: pre;
    word-wrap: normal;
    vertical-align: top;
    color: #1f2328;
  }

  /* Addition lines - GitHub green */
  .diff-view :global(.d2h-ins) {
    background-color: #d1f8d9;
  }

  .diff-view :global(.d2h-ins .d2h-code-linenumber) {
    background-color: #b4f1be;
  }

  .diff-view :global(.d2h-ins .d2h-code-line-prefix) {
    color: #1a7f37;
    background-color: #d1f8d9;
  }

  .diff-view :global(.d2h-ins .d2h-code-line-ctn) {
    background-color: #d1f8d9;
  }

  .diff-view :global(.d2h-ins ins) {
    background-color: #7ee787;
  }

  /* Deletion lines - GitHub red */
  .diff-view :global(.d2h-del) {
    background-color: #ffd7d5;
  }

  .diff-view :global(.d2h-del .d2h-code-linenumber) {
    background-color: #ffc0be;
  }

  .diff-view :global(.d2h-del .d2h-code-line-prefix) {
    color: #cf222e;
    background-color: #ffd7d5;
  }

  .diff-view :global(.d2h-del .d2h-code-line-ctn) {
    background-color: #ffd7d5;
  }

  .diff-view :global(.d2h-del del) {
    background-color: #ff8182;
  }

  /* Context/unchanged lines */
  .diff-view :global(.d2h-cntx) {
    background-color: #ffffff;
  }

  .diff-view :global(.d2h-cntx .d2h-code-linenumber) {
    background-color: #f6f8fa;
  }

  .diff-view :global(.d2h-cntx .d2h-code-line-prefix),
  .diff-view :global(.d2h-cntx .d2h-code-line-ctn) {
    background-color: #ffffff;
  }

  /* Hunk header (@@) - GitHub style */
  .diff-view :global(.d2h-info) {
    background-color: #ddf4ff;
    color: rgba(31, 35, 40, 0.7);
    padding: 6px 10px;
    font-family:
      ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    line-height: 20px;
  }
</style>
