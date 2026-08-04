<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { escapeHtml } from "@rilldata/web-common/lib/i18n";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import { getHasSlackConnection } from "@rilldata/web-common/features/alerts/delivery-tab/notifiers-utils";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import RowsAndColumnsForm from "@rilldata/web-common/features/scheduled-reports/fields/RowsAndColumnsForm.svelte";
  import ScheduleForm from "@rilldata/web-common/features/scheduled-reports/ScheduleForm.svelte";
  import {
    ReportRunAs,
    type ReportValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import type { Readable } from "svelte/store";
  import type { SuperFormErrors } from "sveltekit-superforms/client";
  import Input from "../../components/forms/Input.svelte";
  import Select from "../../components/forms/Select.svelte";
  import Checkbox from "../../components/forms/Checkbox.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import { specHasTabGroups } from "@rilldata/web-common/features/canvas/stores/tab-group";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let formId: string;
  export let data: Readable<ReportValues>;
  export let errors: SuperFormErrors<ReportValues>;
  export let submit: () => void;
  export let enhance;
  // Exactly one of exploreName and canvasName is non-empty; canvasName selects the canvas PDF variant of the form.
  export let exploreName: string;
  export let canvasName: string = "";
  // Canvas state (URL search string) to display instead of the page URL; set when
  // editing a report so the filter bar shows the report's captured state.
  export let canvasStateOverride: string | undefined = undefined;
  export let filters: Filters | undefined = undefined;
  export let timeControls: TimeControls | undefined = undefined;

  const RUN_AS_OPTIONS = [
    {
      value: ReportRunAs.Creator,
      label: m.report_form_run_as_creator(),
      description: m.report_form_run_as_creator_desc(),
    },
    {
      value: ReportRunAs.Recipient,
      label: m.report_form_run_as_recipient(),
      description: m.report_form_run_as_recipient_desc(),
    },
  ];
  const runtimeClient = useRuntimeClient();

  $: selectedRunAsOption = RUN_AS_OPTIONS.find(
    (o) => o.value === $data["webOpenMode"],
  );

  $: hasSlackNotifier = getHasSlackConnection(runtimeClient);

  // Pull the time zone options from the dashboard's spec
  $: exploreSpecQuery = useExploreValidSpec(runtimeClient, exploreName);
  $: canvasQuery = useResource<V1Resource>(
    runtimeClient,
    canvasName,
    ResourceKind.Canvas,
  );
  $: availableTimeZones = canvasName
    ? $canvasQuery.data?.canvas?.state?.validSpec?.timeZones
    : $exploreSpecQuery.data?.explore?.timeZones;

  // Gates the all-tabs option in the canvas PDF variant.
  $: showTabOptions = specHasTabGroups(
    $canvasQuery.data?.canvas?.state?.validSpec?.rows,
  );
  $: canvasFiltersEnabled =
    $canvasQuery.data?.canvas?.state?.validSpec?.filtersEnabled ?? true;

  // Keyboard counterpart of the read-only filter bar's pointer-events guard:
  // its controls remain focusable, so kick focus back out to keep them inoperable.
  function blurFocusedDescendant(event: FocusEvent) {
    if (event.target instanceof HTMLElement) event.target.blur();
  }
</script>

<form
  autocomplete="off"
  class="flex flex-col gap-y-3 w-full"
  id={formId}
  onsubmit={(e) => {
    e.preventDefault();
    submit();
  }}
  use:enhance
>
  <span>{m.report_form_email_recurring()}</span>
  <div class="flex flex-col gap-y-3 w-full h-[600px] overflow-y-scroll">
    <Input
      bind:value={$data["title"]}
      errors={$errors["title"]}
      id="title"
      label={m.report_form_title_label()}
      placeholder={m.report_form_title_placeholder()}
    />
    <Select
      bind:value={$data["webOpenMode"]}
      id="webOpenMode"
      label={m.report_form_run_as()}
      options={RUN_AS_OPTIONS}
      dropdownWidth="w-[400px]"
    />
    {#if selectedRunAsOption}
      <div>
        {selectedRunAsOption.description}
      </div>
    {/if}
    <ScheduleForm {data} {availableTimeZones} />
    {#if canvasName}
      <Select
        value={V1ExportFormat.EXPORT_FORMAT_PDF}
        id="exportFormat"
        label={m.report_form_format()}
        options={[
          {
            value: V1ExportFormat.EXPORT_FORMAT_PDF,
            label: m.report_form_format_pdf(),
          },
        ]}
        disabled
      />
      <Checkbox
        bind:checked={$data["pdfIncludeFilters"]}
        id="pdfIncludeFilters"
        onCheckedChange={(checked) => {
          $data["pdfIncludeFilters"] = Boolean(checked);
        }}
        inverse
        label={m.export_pdf_include_filters()}
      />
      {#if showTabOptions}
        <Checkbox
          bind:checked={$data["pdfAllTabs"]}
          id="pdfAllTabs"
          onCheckedChange={(checked) => {
            $data["pdfAllTabs"] = Boolean(checked);
          }}
          inverse
          label={m.export_pdf_tabs_all()}
        />
      {/if}

      {#if canvasFiltersEnabled}
        <div class="flex flex-col gap-y-3">
          <InputLabel
            label={m.report_form_filters()}
            id="filters"
            capitalize={false}
            hint={m.report_form_filters_readonly_hint()}
          />
          <!-- Read-only display of the canvas filters the report will render with
               (the dashboard's current filters when creating, the report's captured
               state when editing; see ScheduledReportDialog). Report forms must not
               change the underlying dashboard view, so editing is disabled until
               unified filters support in-memory editing. Pointer events are forced
               off on all descendants (the time range pill has no read-only mode and
               the bar re-enables pointer-events internally), which keeps the bar in
               the accessibility tree, unlike inert; blurring on focus covers keyboard
               interaction, which pointer-events does not block. -->
          <CanvasProvider
            {canvasName}
            instanceId={runtimeClient.instanceId}
            isolated
            urlStateOverride={canvasStateOverride}
          >
            <div class="readonly-filter-bar" onfocusin={blurFocusedDescendant}>
              <CanvasFilters {canvasName} maxWidth={820} readOnly />
            </div>
          </CanvasProvider>
        </div>
      {/if}
    {:else}
      <Select
        bind:value={$data["exportFormat"]}
        id="exportFormat"
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
        bind:value={$data["exportLimit"]}
        errors={$errors["exportLimit"]}
        id="exportLimit"
        label={m.report_form_row_limit()}
        optional
        placeholder={m.report_form_row_limit_placeholder()}
      />
      <div class="flex items-center gap-x-1">
        <Checkbox
          bind:checked={$data["exportIncludeHeader"]}
          id="exportIncludeHeader"
          onCheckedChange={(checked) => {
            $data["exportIncludeHeader"] = Boolean(checked);
          }}
          inverse
          disabled={$data["exportFormat"] ===
            V1ExportFormat.EXPORT_FORMAT_PARQUET}
          label={m.report_form_include_metadata()}
        />
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-fg-secondary" style="transform:translateY(-.5px)">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            {m.report_form_metadata_tooltip()}
          </TooltipContent>
        </Tooltip>
      </div>

      {#if filters && timeControls}
        <div class="flex flex-col gap-y-3">
          <InputLabel
            label={m.report_form_filters()}
            id="filters"
            capitalize={false}
          />
          <FiltersForm {filters} {timeControls} side="top" />
        </div>
      {/if}

      <RowsAndColumnsForm
        bind:rows={$data["rows"]}
        bind:columns={$data["columns"]}
        columnErrors={$errors["columns"]}
        {exploreName}
      />
    {/if}

    <MultiInput
      id="emailRecipients"
      label={m.report_form_email_recipients()}
      hint={m.report_form_email_hint()}
      bind:values={$data["emailRecipients"]}
      errors={$errors["emailRecipients"]}
      singular="email"
      plural="emails"
      placeholder={m.report_form_email_placeholder()}
    />
    {#if $hasSlackNotifier.data}
      <FormSection
        bind:enabled={$data["enableSlackNotification"]}
        showSectionToggle
        title={m.report_form_slack_title()}
        padding=""
      >
        <MultiInput
          id="slackChannels"
          label={m.report_form_channels()}
          hint={m.report_form_slack_channels_hint()}
          bind:values={$data["slackChannels"]}
          errors={$errors["slackChannels"]}
          singular="channel"
          plural="channels"
          placeholder={m.alert_form_slack_placeholder()}
        />
        <MultiInput
          id="slackUsers"
          label={m.report_form_slack_users()}
          hint={m.report_form_slack_users_hint()}
          bind:values={$data["slackUsers"]}
          errors={$errors["slackUsers"]}
          singular="user"
          plural="users"
          placeholder={m.report_form_email_placeholder()}
        />
      </FormSection>
    {:else}
      <FormSection title={m.report_form_slack_title()} padding="">
        <svelte:fragment slot="description">
          <span class="text-sm text-fg-secondary">
            {@html m.report_form_slack_not_configured({
              link: `<a href="https://docs.rilldata.com/guides/alerts#configuring-slack-targets" target="_blank">${escapeHtml(m.report_form_docs())}</a>`,
            })}
          </span>
        </svelte:fragment>
      </FormSection>
    {/if}
  </div>
</form>

<style lang="postcss">
  .readonly-filter-bar :global(*) {
    pointer-events: none !important;
  }
</style>
