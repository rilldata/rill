<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import ExportDashboardForm from "./pdf/ExportDashboardForm.svelte";
  import type { PdfExportRunOptions } from "./pdf/types";
  import type { XlsxChartMode } from "./xlsx/types";

  // The export form of a canvas: a PDF/Excel switch (Excel only when the
  // caller offers it), the options shared by both formats, and the chart mode
  // that only Excel has. Actions that apply to PDF alone (scheduling a PDF
  // report) go in the "pdf-actions" slot.
  export let runPdfExport: (opts: PdfExportRunOptions) => Promise<void>;
  export let runXlsxExport:
    | ((opts: PdfExportRunOptions, chartMode: XlsxChartMode) => Promise<void>)
    | null = null;
  // Shown only when the dashboard has tab groups.
  export let showTabOptions = false;
  export let onComplete: () => void = () => {};

  type ExportFormat = "pdf" | "xlsx";
  let format: ExportFormat = "pdf";
  let chartMode: XlsxChartMode = "native";
  // Switching formats while an export runs would mount a fresh form, so the
  // switchers are disabled until the running export completes.
  let exporting = false;

  $: formats = [m.report_form_format_pdf(), m.export_format_excel()];
  $: chartModes = [
    m.export_xlsx_charts_native(),
    m.export_xlsx_charts_images(),
  ];
  $: chartModeHint =
    chartMode === "native"
      ? m.export_xlsx_charts_native_hint()
      : m.export_xlsx_charts_images_hint();
</script>

<div class="flex flex-col gap-y-4">
  {#if runXlsxExport}
    <div class="flex items-center justify-between gap-x-4">
      <span class="text-xs font-semibold text-fg-primary">
        {m.report_form_format()}
      </span>
      <FieldSwitcher
        small
        disabled={exporting}
        fields={formats}
        selected={format === "pdf" ? 0 : 1}
        onClick={(i) => (format = i === 0 ? "pdf" : "xlsx")}
      />
    </div>
  {/if}

  {#if format === "xlsx" && runXlsxExport}
    {@const runXlsx = runXlsxExport}
    <ExportDashboardForm
      bind:exporting
      runExport={(opts: PdfExportRunOptions) => runXlsx(opts, chartMode)}
      {showTabOptions}
      {onComplete}
      idPrefix="xlsx"
      description={m.export_xlsx_description()}
      buttonLabel={m.export_xlsx_button()}
      successMessage={m.export_xlsx_success()}
      failureMessage={m.export_xlsx_failed()}
      progressCopy={{
        preparing: m.export_pdf_rendering_charts(),
        capturing: m.export_pdf_capturing(),
        assembling: m.export_xlsx_building(),
      }}
    >
      <div slot="options" class="flex flex-col gap-y-1">
        <div class="flex items-center justify-between gap-x-4">
          <span class="text-xs font-semibold text-fg-primary">
            {m.export_xlsx_charts_label()}
          </span>
          <FieldSwitcher
            small
            disabled={exporting}
            fields={chartModes}
            selected={chartMode === "native" ? 0 : 1}
            onClick={(i) => (chartMode = i === 0 ? "native" : "image")}
          />
        </div>
        <span class="text-xs text-fg-muted">{chartModeHint}</span>
      </div>
    </ExportDashboardForm>
  {:else}
    <ExportDashboardForm
      bind:exporting
      runExport={runPdfExport}
      {showTabOptions}
      {onComplete}
    />
    <slot name="pdf-actions" />
  {/if}
</div>
