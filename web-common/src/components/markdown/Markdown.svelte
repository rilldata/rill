<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";

  export let content: string;
  export let converter = (c: string) => marked(c);
</script>

<div class="chat-markdown">
  {#await converter(content) then html}
    {@html DOMPurify.sanitize(html)}
  {/await}
</div>

<style lang="postcss">
  :global(.chat-markdown) {
    @apply text-gray-800;
    line-height: 1.6;
  }

  :global(.chat-markdown h1) {
    @apply text-lg font-semibold mb-2;
    margin-top: 1rem;
  }

  :global(.chat-markdown h1:first-child) {
    margin-top: 0;
  }

  :global(.chat-markdown h2) {
    @apply text-base font-semibold mb-2;
    margin-top: 0.75rem;
  }

  :global(.chat-markdown h2:first-child) {
    margin-top: 0;
  }

  :global(.chat-markdown h3) {
    @apply text-sm font-semibold;
    margin-top: 0.75rem;
    margin-bottom: 0.25rem;
  }

  :global(.chat-markdown h3:first-child) {
    margin-top: 0;
  }

  :global(.chat-markdown h4),
  :global(.chat-markdown h5),
  :global(.chat-markdown h6) {
    @apply text-sm font-medium;
    margin-top: 0.5rem;
    margin-bottom: 0.25rem;
  }

  :global(.chat-markdown h4:first-child),
  :global(.chat-markdown h5:first-child),
  :global(.chat-markdown h6:first-child) {
    margin-top: 0;
  }

  :global(.chat-markdown p) {
    @apply text-sm;
    margin-bottom: 0.5rem;
  }

  :global(.chat-markdown p:last-child) {
    margin-bottom: 0;
  }

  :global(.chat-markdown ul) {
    @apply text-sm;
    list-style-type: disc;
    padding-left: 1.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.chat-markdown ul:last-child) {
    margin-bottom: 0;
  }

  :global(.chat-markdown ol) {
    @apply text-sm;
    list-style-type: decimal;
    padding-left: 1.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.chat-markdown ol:last-child) {
    margin-bottom: 0;
  }

  :global(.chat-markdown li) {
    margin-bottom: 0.25rem;
  }

  :global(.chat-markdown li:last-child) {
    margin-bottom: 0;
  }

  :global(.chat-markdown blockquote) {
    @apply border-l-2 border-gray-300 italic text-gray-600 text-sm;
    padding-left: 0.75rem;
    padding-top: 0.25rem;
    padding-bottom: 0.25rem;
    margin-bottom: 0.5rem;
  }

  :global(.chat-markdown code) {
    @apply bg-gray-100 font-mono text-xs;
    padding: 0.125rem 0.25rem;
    border-radius: 0.25rem;
  }

  :global(.chat-markdown pre) {
    @apply bg-gray-100 text-xs;
    padding: 0.5rem;
    border-radius: 0.25rem;
    margin-bottom: 0.5rem;
    overflow-x: auto;
  }

  :global(.chat-markdown pre code) {
    @apply bg-transparent;
    padding: 0;
  }

  :global(.chat-markdown a) {
    @apply text-blue-600 underline;
  }

  :global(.chat-markdown a:hover) {
    @apply text-blue-800;
  }

  :global(.chat-markdown strong) {
    @apply font-semibold;
  }

  :global(.chat-markdown em) {
    @apply italic;
  }

  :global(.chat-markdown hr) {
    @apply border-t border-gray-200;
    margin: 0.75rem 0;
  }

  :global(.chat-markdown table) {
    @apply w-full text-xs;
    border-collapse: collapse;
    margin-bottom: 0.5rem;
  }

  :global(.chat-markdown th) {
    @apply bg-gray-50 border border-gray-200 font-medium;
    padding: 0.25rem 0.5rem;
    text-align: left;
  }

  :global(.chat-markdown td) {
    @apply border border-gray-200;
    padding: 0.25rem 0.5rem;
  }

  :global(.chat-markdown tr:nth-child(even)) {
    @apply bg-gray-50;
  }
</style>
