<script lang="ts">
  import { X } from "lucide-svelte";
  import { fly } from "svelte/transition";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import MetricsViewIcon from "../../../../components/icons/MetricsViewIcon.svelte";
  import type { V1Resource } from "../../../../runtime-client";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import DelayedSpinner from "../../../entity-management/DelayedSpinner.svelte";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../../../entity-management/resource-selectors";
  import Error from "../../core/messages/Error.svelte";

  export let onClose: () => void;

  $: ({ instanceId } = $runtime);

  $: metricsViewsQuery = useFilteredResources(
    instanceId,
    ResourceKind.MetricsView,
  );

  $: metricsViews = $metricsViewsQuery?.data ?? [];
  $: isLoading = $metricsViewsQuery?.isLoading ?? false;
  $: error = $metricsViewsQuery?.error;
  $: errorMessage = error ? error.message || String(error) : undefined;

  function getMetricsViewData(resource: V1Resource | undefined) {
    const name = resource?.meta?.name?.name ?? "";
    const spec = resource?.metricsView?.state?.validSpec;
    const displayName = spec?.displayName;
    const description = spec?.description;
    const measures = spec?.measures ?? [];
    const dimensions = spec?.dimensions ?? [];

    return { name, displayName, description, measures, dimensions };
  }

  // Track which metrics views are collapsed (expanded by default)
  let collapsedViews = new Set<string>();

  function toggleView(name: string) {
    if (collapsedViews.has(name)) {
      collapsedViews.delete(name);
    } else {
      collapsedViews.add(name);
    }
    collapsedViews = collapsedViews;
  }
</script>

<div class="metrics-catalog" transition:fly={{ x: 320, duration: 250 }}>
  <div class="catalog-header">
    <h3 class="catalog-title">Available metrics</h3>
    <IconButton ariaLabel="Close metrics catalog" on:click={onClose}>
      <X size="18px" />
    </IconButton>
  </div>

  <div class="catalog-content">
    {#if isLoading}
      <div class="catalog-loading">
        <DelayedSpinner {isLoading} size="24px" />
      </div>
    {:else if error}
      <Error headline="Failed to load metrics views" error={errorMessage} />
    {:else if metricsViews.length === 0}
      <div class="catalog-empty">
        <p class="catalog-empty-text">No metrics views found</p>
      </div>
    {:else}
      <div class="metrics-list">
        {#each metricsViews as view (view?.meta?.name?.name)}
          {@const { name, displayName, description, measures, dimensions } =
            getMetricsViewData(view)}
          {@const isExpanded = !collapsedViews.has(name)}
          <div class="metrics-view-item">
            <button
              class="metrics-view-header"
              on:click={() => toggleView(name)}
            >
              <svg
                class="expand-icon"
                class:expanded={isExpanded}
                width="14"
                height="14"
                viewBox="0 0 16 16"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M6 4L10 8L6 12"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
              <MetricsViewIcon size="14px" color="#6366F1" />
              <div class="metrics-view-info">
                <span class="metrics-view-name">{displayName || name}</span>
                {#if description}
                  <span class="metrics-view-description">{description}</span>
                {/if}
              </div>
            </button>

            {#if isExpanded}
              <div class="metrics-view-details">
                {#if measures.length > 0}
                  <div class="field-section">
                    <h4 class="field-section-title">
                      Measures ({measures.length})
                    </h4>
                    <ul class="field-list">
                      {#each measures as measure}
                        <li class="field-item">
                          <span class="field-icon measure-icon">Σ</span>
                          <div class="field-info">
                            <div class="field-name-row">
                              <span class="field-name"
                                >{measure.displayName || measure.name}</span
                              >
                            </div>
                            {#if measure.description}
                              <div class="field-description">
                                {measure.description}
                              </div>
                            {/if}
                            {#if measure.expression}
                              <div class="field-expression">
                                {measure.expression}
                              </div>
                            {/if}
                          </div>
                        </li>
                      {/each}
                    </ul>
                  </div>
                {/if}

                {#if dimensions.length > 0}
                  <div class="field-section">
                    <h4 class="field-section-title">
                      Dimensions ({dimensions.length})
                    </h4>
                    <ul class="field-list">
                      {#each dimensions as dimension}
                        <li class="field-item">
                          <span class="field-icon dimension-icon">□</span>
                          <div class="field-info">
                            <div class="field-name-row">
                              <span class="field-name"
                                >{dimension.displayName ||
                                  dimension.name ||
                                  dimension.column}</span
                              >
                            </div>
                            {#if dimension.description}
                              <div class="field-description">
                                {dimension.description}
                              </div>
                            {/if}
                          </div>
                        </li>
                      {/each}
                    </ul>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .metrics-catalog {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--surface);
    border-left: 1px solid var(--border);
    width: 320px;
    flex-shrink: 0;
  }

  .catalog-header {
    padding: 0.75rem 0.75rem 0.5rem 0.75rem;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .catalog-title {
    font-size: 0.75rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .catalog-content {
    flex: 1;
    overflow-y: auto;
    padding: 0 0.5rem 0.5rem 0.5rem;
  }

  .catalog-loading {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 2rem;
  }

  .catalog-empty {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 2rem 1rem;
  }

  .catalog-empty-text {
    color: #6b7280;
    font-size: 0.8125rem;
    text-align: center;
    margin: 0;
  }

  .metrics-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .metrics-view-item {
    background: var(--surface);
  }

  .metrics-view-header {
    position: sticky;
    top: 0;
    z-index: 10;
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.375rem;
    background: var(--surface);
    border: none;
    cursor: pointer;
    text-align: left;
    transition: background-color 0.15s;
    border-radius: 0.25rem;
  }

  .metrics-view-header:hover {
    background-color: #f3f4f6;
  }

  .expand-icon {
    flex-shrink: 0;
    color: #6b7280;
    transition: transform 0.2s;
  }

  .expand-icon.expanded {
    transform: rotate(90deg);
  }

  .metrics-view-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .metrics-view-name {
    font-size: 0.75rem;
    font-weight: 500;
    color: #111827;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .metrics-view-description {
    font-size: 0.625rem;
    color: #6b7280;
    line-height: 1.3;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .metrics-view-details {
    padding: 0.125rem 0.375rem 0.375rem 0.375rem;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .field-section {
    display: flex;
    flex-direction: column;
    gap: 0.0625rem;
  }

  .field-section-title {
    position: sticky;
    top: 1.5rem;
    z-index: 5;
    font-size: 0.625rem;
    font-weight: 600;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    margin: 0 0 0.125rem 0;
    padding: 0.25rem 0 0.125rem 0.25rem;
    background: var(--surface);
  }

  .field-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .field-item {
    display: flex;
    align-items: flex-start;
    gap: 0.375rem;
    padding: 0.125rem 0.25rem;
    border-radius: 0.25rem;
    font-size: 0.6875rem;
    transition: background-color 0.15s;
  }

  .field-item:hover {
    background-color: #f3f4f6;
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
    margin-top: 0.125rem;
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
    gap: 0.125rem;
  }

  .field-name-row {
    display: flex;
    align-items: center;
  }

  .field-name {
    color: #111827;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .field-description {
    color: #6b7280;
    font-size: 0.625rem;
    line-height: 1.3;
  }

  .field-expression {
    color: #6b7280;
    font-family:
      "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", "Source Code Pro",
      Menlo, Consolas, "DejaVu Sans Mono", monospace;
    font-size: 0.625rem;
    line-height: 1.3;
    word-break: break-all;
  }
</style>
