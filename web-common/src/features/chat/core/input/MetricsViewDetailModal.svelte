<script lang="ts">
  import * as Dialog from "../../../../components/dialog";
  import MetricsViewIcon from "../../../../components/icons/MetricsViewIcon.svelte";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
  } from "../../../../runtime-client";

  export let name: string;
  export let displayName: string | undefined;
  export let measures: MetricsViewSpecMeasure[];
  export let dimensions: MetricsViewSpecDimension[];
  export let open: any;
  export let onClose: () => void;
</script>

<Dialog.Root bind:open onOpenChange={(isOpen) => !isOpen && onClose()}>
  <Dialog.Content class="max-w-2xl p-0 gap-0">
    <Dialog.Header class="px-4 py-3 border-b flex-row items-center gap-2">
      <MetricsViewIcon size="18px" color="#6366f1" />
      <Dialog.Title class="text-[15px] font-semibold">
        {displayName || name}
      </Dialog.Title>
    </Dialog.Header>

    <div class="modal-body">
      {#if measures.length > 0}
        <div class="section">
          <h3 class="section-title">Measures ({measures.length})</h3>
          <div class="field-list">
            {#each measures as measure}
              <div class="field-item">
                <div class="field-icon measure-icon">Σ</div>
                <div class="field-info">
                  <div class="field-name">
                    {measure.displayName || measure.name}
                  </div>
                  {#if measure.description}
                    <div class="field-description">{measure.description}</div>
                  {/if}
                  {#if measure.expression}
                    <div class="field-expression">{measure.expression}</div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      {#if dimensions.length > 0}
        <div class="section">
          <h3 class="section-title">Dimensions ({dimensions.length})</h3>
          <div class="field-list">
            {#each dimensions as dimension}
              <div class="field-item">
                <div class="field-icon dimension-icon">□</div>
                <div class="field-info">
                  <div class="field-name">
                    {dimension.displayName ||
                      dimension.name ||
                      dimension.column}
                  </div>
                  {#if dimension.description}
                    <div class="field-description">{dimension.description}</div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style>
  .modal-body {
    overflow-y: auto;
    max-height: 70vh;
    padding: 0.75rem 1rem 1rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .section-title {
    font-size: 0.625rem;
    font-weight: 600;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    margin: 0 0 0.125rem 0;
    padding-left: 0.25rem;
  }

  .field-list {
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .field-item {
    display: flex;
    gap: 0.5rem;
    padding: 0.375rem 0.5rem;
    border-radius: 0.25rem;
    background: transparent;
    transition: background-color 0.15s;
  }

  .field-item:hover {
    background: #f9fafb;
  }

  .field-icon {
    flex-shrink: 0;
    width: 0.875rem;
    height: 0.875rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.625rem;
    font-weight: 600;
    margin-top: 0.25rem;
  }

  .measure-icon {
    color: #3b82f6;
  }

  .dimension-icon {
    color: #6366f1;
  }

  .field-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .field-name {
    font-size: 0.8125rem;
    font-weight: 600;
    color: #111827;
  }

  .field-description {
    font-size: 0.6875rem;
    color: #6b7280;
    line-height: 1.4;
  }

  .field-expression {
    font-family:
      "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", "Source Code Pro",
      Menlo, Consolas, "DejaVu Sans Mono", monospace;
    font-size: 0.6875rem;
    color: #6b7280;
    background: #f9fafb;
    padding: 0.375rem 0.5rem;
    border-radius: 0.25rem;
    word-break: break-all;
    line-height: 1.4;
  }
</style>
