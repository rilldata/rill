<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExportProgress, PdfExportRunOptions } from "./types";
  import type { LocalizedString } from "@inlang/paraglide-js";

  // Surface-agnostic PDF export form (title, options, and action). The caller
  // supplies `runExport` (bound to the canvas or explore orchestrator), so this
  // form is shared across the cloud share modal today and the Rill Developer UI
  // in the future.
  export let runExport: (opts: PdfExportRunOptions) => Promise<void>;
  export let onComplete: () => void = () => {};
  // Shown only when the dashboard has tab groups: whether to export every tab
  // or only each group's active tab.
  export let showTabOptions = false;

  let includeFilters = true;
  let allTabs = true;

  let exporting = false;
  let progressLabel = m.export_pdf_button();

  const PROGRESS_COPY: Record<ExportProgress["phase"], LocalizedString> = {
    preparing: m.export_pdf_rendering_charts(),
    capturing: m.export_pdf_capturing(),
    assembling: m.export_pdf_building(),
  };

  async function onExport() {
    // Guard against re-entry: the button shows a spinner while exporting but
    // stays clickable, and overlapping exports share one capture header.
    if (exporting) return;
    exporting = true;
    progressLabel = PROGRESS_COPY.preparing;
    try {
      await runExport({
        includeFilters,
        allTabs,
        onProgress: ({ phase }) => {
          progressLabel = PROGRESS_COPY[phase];
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: m.export_pdf_success(),
      });
      onComplete();
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(e) || m.export_pdf_failed(),
      });
    } finally {
      exporting = false;
      progressLabel = m.export_pdf_button();
    }
  }
</script>

<div class="flex flex-col gap-y-4">
  <h3 class="text-xs text-fg-primary font-normal">
    {m.export_pdf_description()}
  </h3>

  <Checkbox
    id="pdf-include-filters"
    bind:checked={includeFilters}
    label={m.export_pdf_include_filters()}
  />

  {#if showTabOptions}
    <Checkbox
      id="pdf-all-tabs"
      bind:checked={allTabs}
      label={m.export_pdf_tabs_all()}
    />
  {/if}

  <Button
    type="primary"
    loading={exporting}
    loadingCopy={progressLabel}
    onClick={onExport}
  >
    {m.export_pdf_button()}
  </Button>
</div>

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
