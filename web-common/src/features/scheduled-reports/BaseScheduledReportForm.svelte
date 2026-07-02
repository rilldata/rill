<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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

  export let formId: string;
  export let data: Readable<ReportValues>;
  export let errors: SuperFormErrors<ReportValues>;
  export let submit: () => void;
  export let enhance;
  export let exploreName: string;
  export let filters: Filters;
  export let timeControls: TimeControls;

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
    <ScheduleForm {data} {exploreName} />
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

    <div class="flex flex-col gap-y-3">
      <InputLabel
        label={m.report_form_filters()}
        id="filters"
        capitalize={false}
      />
      <FiltersForm {filters} {timeControls} side="top" />
    </div>

    <RowsAndColumnsForm
      bind:rows={$data["rows"]}
      bind:columns={$data["columns"]}
      columnErrors={$errors["columns"]}
      {exploreName}
    />

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
              link: `<a href="https://docs.rilldata.com/guides/alerts#configuring-slack-targets" target="_blank">${m.report_form_docs()}</a>`,
            })}
          </span>
        </svelte:fragment>
      </FormSection>
    {/if}
  </div>
</form>
