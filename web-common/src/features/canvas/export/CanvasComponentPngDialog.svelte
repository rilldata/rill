<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import { hideBorder } from "@rilldata/web-common/features/canvas/layout-util";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
  import { buildExportFilename } from "@rilldata/web-common/features/exports/filename";
  import { downloadNodeAsPng } from "@rilldata/web-common/features/exports/png/export-png";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter";

  // Preview-and-download dialog for a single canvas component, the canvas
  // counterpart of explore's ScreenshotContainer. The component is rendered a
  // second time off the same store, at the size it occupies on the canvas, and
  // framed with the dashboard's title, time range and filters so the image
  // stands on its own once shared. The frame is scaled down to fit the dialog
  // but captured at its natural size.
  let {
    open = $bindable(false),
    component,
    width,
    height,
  }: {
    open?: boolean;
    component: BaseCanvasComponent;
    // The component card's on-screen size in CSS pixels.
    width: number;
    height: number;
  } = $props();

  // Horizontal padding + border of the frame around the card.
  const FRAME_INSET_PX = 2 * 16 + 2;
  // The dialog's own horizontal padding.
  const DIALOG_INSET_PX = 2 * 24;

  // Embedded dashboards live inside a customer's product, so the exported
  // image should not carry Rill branding there.
  const isEmbedded = EmbedStore.isEmbedded();

  let {
    theme,
    titleStore,
    expressionFilterManager,
    timeManager: {
      state: {
        interval: intervalStore,
        grainStore,
        timeZoneStore,
        comparisonIntervalStore,
        showTimeComparisonStore,
      },
    },
  } = $derived(component.parent);
  let spec = $derived(component.specStore);

  let includeHeader = $state(true);
  let downloading = $state(false);
  let captureNode = $state<HTMLDivElement>();
  let availableWidth = $state(0);
  let frameHeight = $state(0);

  let frameWidth = $derived(width + FRAME_INSET_PX);
  let scale = $derived(
    availableWidth > 0 ? Math.min(1, availableWidth / frameWidth) : 1,
  );
  let allowBorder = $derived(!hideBorder.has(component.type));

  // Exact, resolved range (e.g. "Jan 1 – Jan 7, 2024"), never the relative alias.
  let formattedTimeRange = $derived(
    $intervalStore ? prettyFormatTimeRange($intervalStore, $grainStore) : "",
  );
  let formattedComparisonRange = $derived(
    $showTimeComparisonStore && $comparisonIntervalStore
      ? prettyFormatTimeRange($comparisonIntervalStore, $grainStore)
      : "",
  );
  let exportTitle = $derived($spec?.title || $titleStore || component.id);
  const generatedTime = new Date().toISOString();

  async function download() {
    if (!captureNode || downloading) return;
    downloading = true;
    try {
      await downloadNodeAsPng(
        captureNode,
        buildExportFilename(exportTitle, "png"),
      );
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(e) || m.canvas_png_export_failed(),
      });
    } finally {
      downloading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <!-- Sized to the frame so a narrow card doesn't float in an empty dialog,
       within bounds that keep the controls usable and the dialog on screen. -->
  <Dialog.Content
    class="flex flex-col gap-y-4"
    style="max-width: min(56rem, max(32rem, {frameWidth + DIALOG_INSET_PX}px))"
  >
    <Dialog.Header>
      <Dialog.Title>{m.dashboard_download_as_png()}</Dialog.Title>
    </Dialog.Header>

    <ThemeProvider theme={$theme} applyLayout={false}>
      <div class="w-full overflow-hidden" bind:clientWidth={availableWidth}>
        <!-- Reserves the scaled frame's footprint in the dialog's layout. -->
        <div
          style:width="{frameWidth * scale}px"
          style:height="{frameHeight * scale}px"
        >
          <!-- The scale lives on this wrapper, not on the capture node: html-to-image
               copies the node's computed transform onto its clone, but ignores
               the ancestors', so the capture comes out at natural size. -->
          <div
            style:width="{frameWidth}px"
            style:transform="scale({scale})"
            style:transform-origin="top left"
          >
            <div
              bind:this={captureNode}
              bind:clientHeight={frameHeight}
              id="canvas-png-export-frame"
              class="flex flex-col gap-y-3 p-4 bg-surface-background border rounded-md"
            >
              {#if includeHeader}
                <header class="flex flex-col gap-y-0.5">
                  <h2 class="text-base font-semibold text-fg-base">
                    {$titleStore}
                  </h2>
                  {#if formattedTimeRange}
                    <div class="text-sm text-fg-secondary">
                      {formattedTimeRange}
                      {#if formattedComparisonRange}
                        <span>{m.time_vs()} {formattedComparisonRange}</span>
                      {/if}
                      <span class="text-fg-muted">· {$timeZoneStore}</span>
                    </div>
                  {/if}
                </header>

                {#if expressionFilterManager.hasSomeFilter}
                  <ReadonlyExpressionFilters
                    {expressionFilterManager}
                    ariaLabel={undefined}
                  />
                {/if}
              {/if}

              <!-- Mirrors the canvas's row and item wrappers, whose container
                   queries components size themselves against. -->
              <div class="canvas-container-frame w-full">
                <div
                  class="component-container-frame"
                  style:width="{width}px"
                  style:height="{height}px"
                >
                  <article
                    class="size-full flex flex-col bg-surface-card overflow-hidden rounded-sm"
                    class:bordered={allowBorder}
                  >
                    <component.component {component} editable={false} />
                  </article>
                </div>
              </div>

              <footer
                class="flex items-center justify-between text-xs text-fg-muted"
              >
                {#if !isEmbedded}
                  <!-- i18n-ignore: standalone product name -->
                  <span>Rill</span>
                {/if}
                <span class="ml-auto">
                  {m.dashboard_generated({ time: generatedTime })}
                </span>
              </footer>
            </div>
          </div>
        </div>
      </div>
    </ThemeProvider>

    <Checkbox
      id="png-include-header"
      bind:checked={includeHeader}
      label={m.canvas_png_include_header()}
    />

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>
        {m.dashboard_cancel()}
      </Button>
      <Button
        type="primary"
        loading={downloading}
        loadingCopy={m.dashboard_generating()}
        onClick={download}
      >
        {m.dashboard_download_png()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .canvas-container-frame {
    container-type: inline-size;
    container-name: canvas-container;
  }

  .component-container-frame {
    container-type: inline-size;
    container-name: component-container;
  }

  .bordered {
    @apply outline outline-[1px] outline-border shadow-sm;
  }

  /* Scrollbars on overflowing components (tables, legends) would be rasterized
     into the capture. Applied per element because html-to-image inlines each
     element's computed styles into its clone: computed scrollbar-width survives
     the cloning, whereas ::-webkit-scrollbar stylesheet rules do not. */
  #canvas-png-export-frame :global(*) {
    scrollbar-width: none;
  }
</style>
