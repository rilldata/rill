<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import type { ExportProgress, PdfExportRunOptions } from "./types";

  // Surface-agnostic PDF export options + action. The caller supplies `runExport`
  // (bound to the canvas or explore orchestrator), so this form is shared across
  // the cloud share modal today and the Rill Developer UI in the future.
  export let runExport: (opts: PdfExportRunOptions) => Promise<void>;
  export let onComplete: () => void = () => {};

  let includeFilters = true;

  let exporting = false;
  let progressLabel = "Export PDF";

  const PROGRESS_COPY: Record<ExportProgress["phase"], string> = {
    preparing: "Rendering charts…",
    capturing: "Capturing dashboard…",
    assembling: "Building PDF…",
  };

  async function onExport() {
    exporting = true;
    progressLabel = PROGRESS_COPY.preparing;
    try {
      await runExport({
        includeFilters,
        onProgress: ({ phase }) => {
          progressLabel = PROGRESS_COPY[phase];
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Dashboard exported as PDF",
      });
      onComplete();
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(e) || "Failed to export PDF",
      });
    } finally {
      exporting = false;
      progressLabel = "Export PDF";
    }
  }
</script>

<div class="flex flex-col gap-y-4">
  <Checkbox
    id="pdf-include-filters"
    bind:checked={includeFilters}
    label="Include filters"
  />

  <Button
    type="primary"
    loading={exporting}
    loadingCopy={progressLabel}
    onClick={onExport}
  >
    Export PDF
  </Button>
</div>
