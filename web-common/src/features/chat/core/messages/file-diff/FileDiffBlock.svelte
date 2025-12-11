<!--
  Renders a file diff block with collapsible tool call header.
  Shows the diff visualization with expandable request/response details.
-->
<script lang="ts">
  import { html } from "diff2html";
  import "diff2html/bundles/css/diff2html.min.css";
  import DOMPurify from "dompurify";
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { FileDiffBlock } from "./file-diff-block";

  export let block: FileDiffBlock;
  export let tools: V1Tool[] | undefined = undefined;

  // Generate and sanitize diff HTML
  $: diffHtml = block.diff
    ? DOMPurify.sanitize(
        html(block.diff, {
          drawFileList: false,
          outputFormat: "line-by-line",
          matching: "lines",
        }),
      )
    : "";
</script>

<div class="file-diff-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />

  <div class="diff-container">
    <div class="diff-header">
      <a href="/files{block.filePath}" class="file-path-link">
        {block.filePath}
      </a>
      {#if block.isNewFile}
        <span class="new-badge">new</span>
      {/if}
    </div>
    {#if diffHtml}
      <div class="diff-view">
        {@html diffHtml}
      </div>
    {:else}
      <div class="no-changes-message">No changes detected</div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .file-diff-block {
    @apply w-full max-w-full self-start;
  }

  /* Diff container */
  .diff-container {
    @apply border border-gray-200 rounded-md overflow-hidden;
  }

  .diff-header {
    @apply flex items-center gap-2 px-3 py-1.5;
    @apply text-xs border-b border-gray-200;
    background-color: #f6f8fa;
  }

  .file-path-link {
    @apply text-gray-600 font-mono;
  }

  .file-path-link:hover {
    @apply text-gray-800 underline;
  }

  .new-badge {
    @apply text-[0.5rem] px-1 py-0.5 rounded;
    @apply bg-green-100 text-green-700 font-medium;
  }

  .no-changes-message {
    @apply p-3 text-xs text-gray-500 italic;
  }

  .diff-view {
    @apply overflow-x-auto;
  }

  /* Structural overrides for diff2html */
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
