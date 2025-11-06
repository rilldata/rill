<script lang="ts">
  import CaretDownIcon from "../../../../components/icons/CaretDownIcon.svelte";
  import ChevronRight from "../../../../components/icons/ChevronRight.svelte";

  export let toolCall: any;
  export let toolResult: any = null;
  export let isExpanded: boolean = false;
  export let onToggle: () => void;

  function formatJson(obj: any): string {
    return JSON.stringify(obj, null, 2);
  }
</script>

<div class="tool-container">
  <button class="tool-header" on:click={onToggle}>
    <div class="tool-icon">
      {#if isExpanded}
        <CaretDownIcon size="16" />
      {:else}
        <ChevronRight size="16" />
      {/if}
    </div>
    <div class="tool-name">
      {toolCall.name || "Unknown Tool"}
    </div>
  </button>

  {#if isExpanded}
    <div class="tool-content">
      <div class="tool-section">
        <div class="tool-section-title">Request</div>
        <div class="tool-section-content">
          <pre class="tool-json">{formatJson(toolCall.input || {})}</pre>
        </div>
      </div>

      {#if toolResult}
        <div class="tool-section">
          <div class="tool-section-title">
            {toolResult.isError ? "Error" : "Response"}
          </div>
          <div class="tool-section-content">
            <pre class="tool-json">{toolResult.content || ""}</pre>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  @reference "tailwindcss";

  @reference "tailwindcss";

  .tool-container {
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    background: #fafafa;
    width: 100%;
  }

  .tool-header {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem;
    background: none;
    border: none;
    cursor: pointer;
    font-size: 0.875rem;
    transition: background-color 0.15s ease;
  }

  .tool-header:hover {
    background: #f3f4f6;
  }

  .tool-icon {
    color: #6b7280;
    display: flex;
    align-items: center;
  }

  .tool-name {
    font-weight: 500;
    color: #374151;
    flex: 1;
    text-align: left;
  }

  .tool-content {
    border-top: 1px solid #e5e7eb;
    background: #ffffff;
    border-radius: 0 0 0.5rem 0.5rem;
  }

  .tool-section {
    padding: 0.5rem;
  }

  .tool-section:not(:last-child) {
    border-bottom: 1px solid #f3f4f6;
  }

  .tool-section-title {
    font-size: 0.625rem;
    font-weight: 600;
    color: #6b7280;
    margin-bottom: 0.375rem;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    display: flex;
    align-items: center;
  }

  .tool-section-content {
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.375rem;
    overflow: hidden;
  }

  .tool-json {
    font-family:
      "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas,
      "Courier New", monospace;
    font-size: 0.75rem;
    line-height: 1.4;
    color: #374151;
    padding: 0.5rem;
    margin: 0;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
