<script lang="ts">
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";

  // Options for a data export file, shared by scheduled reports and alert attachments.
  let {
    format = $bindable(),
    limit = $bindable(),
    includeHeader = $bindable(),
    limitErrors = undefined,
  }: {
    format: V1ExportFormat;
    limit: string;
    includeHeader: boolean;
    limitErrors?: string[];
  } = $props();
</script>

<Select
  bind:value={format}
  id="exportFormat"
  onChange={(value) => {
    // Parquet files have no header, so clear the option instead of leaving it hidden behind the disabled checkbox
    if (value === V1ExportFormat.EXPORT_FORMAT_PARQUET) includeHeader = false;
  }}
  label={m.report_form_format()}
  options={[
    {
      value: V1ExportFormat.EXPORT_FORMAT_CSV,
      label: m.report_form_format_csv(),
    },
    {
      value: V1ExportFormat.EXPORT_FORMAT_PARQUET,
      label: m.report_form_format_parquet(),
    },
    {
      value: V1ExportFormat.EXPORT_FORMAT_XLSX,
      label: m.report_form_format_xlsx(),
    },
  ]}
/>
<Input
  bind:value={limit}
  errors={limitErrors}
  id="exportLimit"
  label={m.report_form_row_limit()}
  optional
  placeholder={m.report_form_row_limit_placeholder()}
/>
<Checkbox
  bind:checked={includeHeader}
  id="exportIncludeHeader"
  onCheckedChange={(checked) => {
    includeHeader = Boolean(checked);
  }}
  inverse
  disabled={format === V1ExportFormat.EXPORT_FORMAT_PARQUET}
  label={m.report_form_include_metadata()}
  hint={m.report_form_metadata_tooltip()}
/>
