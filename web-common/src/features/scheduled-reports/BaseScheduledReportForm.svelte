<script lang="ts">
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import { getHasSlackConnection } from "@rilldata/web-common/features/alerts/delivery-tab/notifiers-utils";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import RowsAndColumnsForm from "@rilldata/web-common/features/scheduled-reports/fields/RowsAndColumnsForm.svelte";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import type { Readable } from "svelte/store";
  import type { SuperFormErrors } from "sveltekit-superforms/client";
  import Input from "../../components/forms/Input.svelte";
  import Select from "../../components/forms/Select.svelte";
  import Checkbox from "../../components/forms/Checkbox.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import { makeTimeZoneOptions, ReportFrequency } from "./time-utils";

  export let formId: string;
  export let data: Readable<ReportValues>;
  export let errors: SuperFormErrors<ReportValues>;
  export let submit: () => void;
  export let enhance;
  export let exploreName: string;
  export let filters: Filters;
  export let timeControls: TimeControls;

  $: ({ instanceId } = $runtime);

  // Pull the time zone options from the dashboard's spec
  $: exploreSpec = useExploreValidSpec(instanceId, exploreName);
  $: availableTimeZones = $exploreSpec.data?.explore?.timeZones;
  $: timeZoneOptions = makeTimeZoneOptions(availableTimeZones);
  $: hasSlackNotifier = getHasSlackConnection(instanceId);
</script>

<form
  autocomplete="off"
  class="flex flex-col gap-y-3 w-full"
  id={formId}
  on:submit|preventDefault={submit}
  use:enhance
>
  <span>Email recurring exports to recipients.</span>

  <div class="flex flex-col gap-y-3 w-full h-[600px] overflow-y-scroll">
    <Input
      bind:value={$data["title"]}
      errors={$errors["title"]}
      id="title"
      label="Report title"
      placeholder="My report"
    />
    <div class="flex gap-x-1">
      <Select
        bind:value={$data["frequency"]}
        id="frequency"
        label="Frequency"
        options={["Daily", "Weekdays", "Weekly", "Monthly"].map(
          (frequency) => ({
            value: frequency,
            label: frequency,
          }),
        )}
      />
      {#if $data["frequency"] === ReportFrequency.Weekly}
        <Select
          bind:value={$data["dayOfWeek"]}
          id="dayOfWeek"
          label="Day"
          options={[
            "Monday",
            "Tuesday",
            "Wednesday",
            "Thursday",
            "Friday",
            "Saturday",
            "Sunday",
          ].map((day) => ({
            value: day,
            label: day,
          }))}
        />
      {/if}
      {#if $data["frequency"] === ReportFrequency.Monthly}
        <Select
          value={"1"}
          id="dayOfMonth"
          label="Day"
          options={[{ value: "1", label: "First day" }]}
          disabled
        />
      {/if}
      <TimePicker bind:value={$data["timeOfDay"]} id="timeOfDay" label="Time" />
      <Select
        bind:value={$data["timeZone"]}
        id="timeZone"
        label="Time zone"
        options={timeZoneOptions}
      />
    </div>
    <Select
      bind:value={$data["exportFormat"]}
      id="exportFormat"
      label="Format"
      options={[
        { value: V1ExportFormat.EXPORT_FORMAT_CSV, label: "CSV" },
        { value: V1ExportFormat.EXPORT_FORMAT_PARQUET, label: "Parquet" },
        { value: V1ExportFormat.EXPORT_FORMAT_XLSX, label: "XLSX" },
      ]}
    />
    <Input
      bind:value={$data["exportLimit"]}
      errors={$errors["exportLimit"]}
      id="exportLimit"
      label="Row limit"
      optional
      placeholder="1000"
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
        label="Include metadata"
      />
      <Tooltip location="right" alignment="middle" distance={8}>
        <div class="text-gray-500" style="transform:translateY(-.5px)">
          <InfoCircle size="13px" />
        </div>
        <TooltipContent maxWidth="400px" slot="tooltip-content">
          Adds a header to the file that includes filters, time range, and other
          metadata.
        </TooltipContent>
      </Tooltip>
    </div>

    <FormSection title="Filters" padding="">
      <FiltersForm {filters} {timeControls} maxWidth={750} side="top" />
    </FormSection>

    <RowsAndColumnsForm
      bind:rows={$data["rows"]}
      bind:columns={$data["columns"]}
      {instanceId}
      {exploreName}
    />

    <MultiInput
      id="emailRecipients"
      label="Email Recipients"
      hint="Recipients will receive different views based on their security policy.
        Recipients without project access can only download the report."
      bind:values={$data["emailRecipients"]}
      errors={$errors["emailRecipients"]}
      singular="email"
      plural="emails"
      placeholder="Enter an email address"
    />
    {#if $hasSlackNotifier.data}
      <FormSection
        bind:enabled={$data["enableSlackNotification"]}
        showSectionToggle
        title="Slack notifications"
        padding=""
      >
        <MultiInput
          id="slackChannels"
          label="Channels"
          hint="We’ll send alerts directly to these channels."
          bind:values={$data["slackChannels"]}
          errors={$errors["slackChannels"]}
          singular="channel"
          plural="channels"
          placeholder="# Enter a Slack channel name"
        />
        <MultiInput
          id="slackUsers"
          label="Users"
          hint="We’ll alert them with direct messages in Slack."
          bind:values={$data["slackUsers"]}
          errors={$errors["slackUsers"]}
          singular="user"
          plural="users"
          placeholder="Enter an email address"
        />
      </FormSection>
    {:else}
      <FormSection title="Slack notifications" padding="">
        <svelte:fragment slot="description">
          <span class="text-sm text-slate-600">
            Slack has not been configured for this project. Read the <a
              href="https://docs.rilldata.com/explore/alerts/slack"
              target="_blank"
            >
              docs
            </a> to learn more.
          </span>
        </svelte:fragment>
      </FormSection>
    {/if}
  </div>
</form>
