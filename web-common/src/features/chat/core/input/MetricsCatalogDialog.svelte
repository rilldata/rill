<script lang="ts">
  import * as Dialog from "../../../../components/dialog";
  import MetricsViewIcon from "../../../../components/icons/MetricsViewIcon.svelte";
  import type { V1Resource } from "../../../../runtime-client";

  export let open: boolean;
  export let metricsViews: Array<V1Resource | undefined>;
  export let onClose: () => void;

  let selectedViewIndex = 0;

  function getMetricsData(view: V1Resource | undefined) {
    const name = view?.meta?.name?.name ?? "";
    const spec = view?.metricsView?.state?.validSpec;
    const displayName = spec?.displayName;
    const description = spec?.description;
    const measures = spec?.measures ?? [];
    const dimensions = spec?.dimensions ?? [];
    return { name, displayName, description, measures, dimensions };
  }

  $: selectedView = metricsViews[selectedViewIndex];
  $: selectedData = getMetricsData(selectedView);
</script>

<Dialog.Root bind:open onOpenChange={(isOpen) => !isOpen && onClose()}>
  <Dialog.Content class="max-w-4xl p-0 gap-0 h-[75vh]">
    <Dialog.Header class="px-3 py-2.5 border-b flex-col items-start gap-0.5">
      <Dialog.Title class="text-sm font-semibold">
        Available Metrics
      </Dialog.Title>
      <p class="dialog-subtitle">Explore the data you can query and analyze</p>
    </Dialog.Header>

    <div class="catalog-container">
      <!-- Sidebar -->
      <div class="catalog-sidebar">
        {#each metricsViews as view, index}
          {@const { name, displayName } = getMetricsData(view)}
          <button
            class="sidebar-item"
            class:selected={selectedViewIndex === index}
            on:click={() => (selectedViewIndex = index)}
          >
            <MetricsViewIcon
              size="16px"
              color={selectedViewIndex === index ? "#6366f1" : "#9ca3af"}
            />
            <div class="sidebar-item-name">{displayName || name}</div>
          </button>
        {/each}
      </div>

      <!-- Main Content -->
      <div class="catalog-main">
        <div class="catalog-content">
          {#if selectedData.measures.length > 0}
            <div class="section">
              <h3 class="section-title">
                Measures ({selectedData.measures.length})
              </h3>
              <div class="field-list">
                {#each selectedData.measures as measure}
                  <div class="field-item">
                    <div class="field-icon measure-icon">Σ</div>
                    <div class="field-info">
                      <div class="field-name">
                        {measure.displayName || measure.name}
                      </div>
                      {#if measure.description}
                        <div class="field-description">
                          {measure.description}
                        </div>
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

          {#if selectedData.dimensions.length > 0}
            <div class="section">
              <h3 class="section-title">
                Dimensions ({selectedData.dimensions.length})
              </h3>
              <div class="field-list">
                {#each selectedData.dimensions as dimension}
                  <div class="field-item">
                    <div class="field-icon dimension-icon">□</div>
                    <div class="field-info">
                      <div class="field-name">
                        {dimension.displayName ||
                          dimension.name ||
                          dimension.column}
                      </div>
                      {#if dimension.description}
                        <div class="field-description">
                          {dimension.description}
                        </div>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<style>
  .dialog-subtitle {
    font-size: 0.75rem;
    color: #6b7280;
    font-weight: 400;
    margin: 0;
  }

  .catalog-container {
    display: flex;
    height: 100%;
    overflow: hidden;
  }

  .catalog-sidebar {
    width: 220px;
    border-right: 1px solid #e5e7eb;
    overflow-y: auto;
    background: #f9fafb;
    flex-shrink: 0;
  }

  .sidebar-item {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border: none;
    background: none;
    cursor: pointer;
    text-align: left;
    transition: all 0.15s;
    border-left: 3px solid transparent;
  }

  .sidebar-item:hover {
    background: #f3f4f6;
  }

  .sidebar-item.selected {
    background: #eef2ff;
    border-left-color: #6366f1;
  }

  .sidebar-item.selected .sidebar-item-name {
    font-weight: 600;
    color: #111827;
  }

  .sidebar-item-name {
    font-size: 0.8125rem;
    font-weight: 500;
    color: #111827;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
  }

  .catalog-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: white;
  }

  .catalog-content {
    flex: 1;
    overflow-y: auto;
    padding: 0.75rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .section-title {
    font-size: 0.6875rem;
    font-weight: 600;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding-left: 0;
    margin-bottom: 0.125rem;
  }

  .field-list {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .field-item {
    display: flex;
    gap: 0.5rem;
    align-items: flex-start;
    padding: 0;
  }

  .field-icon {
    flex-shrink: 0;
    width: 1rem;
    height: 1rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.6875rem;
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
    gap: 0.1875rem;
  }

  .field-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: #111827;
    line-height: 1.3;
  }

  .field-description {
    font-size: 0.75rem;
    color: #6b7280;
    line-height: 1.4;
  }

  .field-expression {
    font-family:
      "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", "Source Code Pro",
      Menlo, Consolas, "DejaVu Sans Mono", monospace;
    font-size: 0.6875rem;
    color: #52525b;
    background: #fafafa;
    border: 1px solid #e5e7eb;
    padding: 0.1875rem 0.375rem;
    border-radius: 0.1875rem;
    line-height: 1.3;
    word-break: break-all;
    display: inline-block;
    width: fit-content;
  }
</style>
