<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExportProgress, PdfExportRunOptions } from "./types";

  // Dashboard export form (description, options, and action) shared by the
  // PDF and Excel exports, in the cloud share modal today and the Rill
  // Developer UI in the future. The caller supplies `runExport` (bound to the
  // canvas orchestrator of its format) and the copy; the defaults are the PDF
  // copy. Format-specific options go in the "options" slot.
  export let runExport: (opts: PdfExportRunOptions) => Promise<void>;
  export let onComplete: () => void = () => {};
  // Shown only when the dashboard has tab groups: whether to export every tab
  // or only each group's active tab.
  export let showTabOptions = false;
  export let description: string = m.export_pdf_description();
  export let buttonLabel: string = m.export_pdf_button();
  export let successMessage: string = m.export_pdf_success();
  export let failureMessage: string = m.export_pdf_failed();
  export let progressCopy: Record<ExportProgress["phase"], string> = {
    preparing: m.export_pdf_rendering_charts(),
    capturing: m.export_pdf_capturing(),
    assembling: m.export_pdf_building(),
  };
  // Prefix of the form control ids, unique per mounted form.
  export let idPrefix = "pdf";
  // True while an export runs. Bindable, so a parent that can swap this form
  // for another (e.g. a format switch) can keep that from interrupting it.
  export let exporting = false;

  let includeFilters = true;
  let allTabs = true;

  let progressLabel: string = buttonLabel;

  async function onExport() {
    // Guard against re-entry: the button shows a spinner while exporting but
    // stays clickable, and overlapping exports share one capture header.
    if (exporting) return;
    exporting = true;
    progressLabel = progressCopy.preparing;
    try {
      await runExport({
        includeFilters,
        allTabs,
        onProgress: ({ phase }) => {
          progressLabel = progressCopy[phase];
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: successMessage,
      });
      onComplete();
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(e) || failureMessage,
      });
    } finally {
      exporting = false;
      progressLabel = buttonLabel;
    }
  }
</script>

<div class="flex flex-col gap-y-4">
  <h3 class="text-xs text-fg-primary font-normal">
    {description}
  </h3>

  <Checkbox
    id="{idPrefix}-include-filters"
    bind:checked={includeFilters}
    label={m.export_pdf_include_filters()}
  />

  {#if showTabOptions}
    <Checkbox
      id="{idPrefix}-all-tabs"
      bind:checked={allTabs}
      label={m.export_pdf_tabs_all()}
    />
  {/if}

  <slot name="options" />

  <Button
    type="primary"
    loading={exporting}
    loadingCopy={progressLabel}
    onClick={onExport}
  >
    {buttonLabel}
  </Button>
</div>

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
