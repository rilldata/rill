<script lang="ts" context="module">
  import type {
    V1Query,
    V1ReportSpec,
  } from "@rilldata/web-common/runtime-client";

  export type CreateReportProps = {
    mode: "create";
    query: V1Query;
    exploreName: string;
  };

  export type EditReportProps = {
    mode: "edit";
    reportName: string;
    reportSpec: V1ReportSpec;
  };
</script>

<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceCreateReport,
    createAdminServiceEditReport,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getHasSlackConnection } from "@rilldata/web-common/features/alerts/delivery-tab/notifiers-utils.ts";
  import { getPivotExportQuery } from "@rilldata/web-common/features/dashboards/pivot/pivot-export.ts";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import ScheduleForm from "@rilldata/web-common/features/scheduled-reports/ScheduleForm.svelte";
  import { convertFormValuesToCronExpression } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
  import {
    getDashboardNameFromReport,
    getExistingReportInitialFormValues,
    getNewReportInitialFormValues,
    getQueryNameFromQuery,
    type ReportValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
    type V1MetricsViewAggregationRequest,
    type V1ReportSpecAnnotations,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { get } from "svelte/store";
  import { defaults, superForm } from "sveltekit-superforms";
  import { type ValidationAdapter, yup } from "sveltekit-superforms/adapters";
  import Button from "web-common/src/components/button/Button.svelte";
  import { array, boolean, object, string } from "yup";

  export let props: CreateReportProps | EditReportProps;
  export let organization: string;
  export let project: string;

  const user = createAdminServiceGetCurrentUser();
  const FORM_ID = "scheduled-report-form";

  $: ({ instanceId } = $runtime);
  const stateManagers = getStateManagers();
  const { dashboardStore } = stateManagers;

  $: reportName = props.mode === "create" ? "" : props.reportName;

  $: exploreName =
    props.mode === "create"
      ? props.exploreName
      : getDashboardNameFromReport(props.reportSpec);

  $: mutation =
    props.mode === "create"
      ? createAdminServiceCreateReport()
      : createAdminServiceEditReport();

  $: queryName =
    props.mode === "create"
      ? getQueryNameFromQuery(props.query)
      : props.reportSpec.queryName;
  $: aggregationRequest = (
    props.mode === "create"
      ? props.query.metricsViewAggregationRequest
      : JSON.parse(props.reportSpec.queryArgsJson || "{}")
  ) as V1MetricsViewAggregationRequest;

  const schema = yup(
    object({
      title: string().required("Required"),
      emailRecipients: array().of(string().email("Invalid email")),
      enableSlackNotification: boolean(), // Needed to get the type for validation
      slackChannels: array().of(string()),
      slackUsers: array().of(string().email("Invalid email")),
    }).test(
      "at-least-one-recipient",
      "At least one email recipient, slack user, or slack channel is required",
      function (value) {
        // Check if at least one array has non-empty values
        const hasEmailRecipients = value.emailRecipients
          ? value.emailRecipients.filter(Boolean).length > 0
          : false;
        if (!value.enableSlackNotification) return hasEmailRecipients;

        const hasSlackUsers = value.slackUsers
          ? value.slackUsers.filter(Boolean).length > 0
          : false;
        const hasSlackChannels = value.slackChannels
          ? value.slackChannels.filter(Boolean).length > 0
          : false;

        return hasEmailRecipients || hasSlackUsers || hasSlackChannels;
      },
    ),
  ) as ValidationAdapter<ReportValues>;

  $: initialValues =
    props.mode === "create"
      ? getNewReportInitialFormValues(
          $user.data?.user?.email,
          aggregationRequest,
        )
      : getExistingReportInitialFormValues(
          props.reportSpec,
          $user.data?.user?.email,
          aggregationRequest,
        );

  $: hasSlackNotifier = getHasSlackConnection(instanceId);

  $: ({ form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      id: FORM_ID,
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        return handleSubmit(values);
      },
      // We need to run the 1st validation only after a submit.
      // But successive validations should be on input.
      // Here, "auto" achieves this.
      validationMethod: "auto",
      invalidateAll: false,
    },
  ));

  $: generalErrors = $errors._errors?.[0] ?? $mutation.error?.message;
  $: console.log($errors);

  async function handleSubmit(values: ReportValues) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
      values.dayOfMonth,
    );
    const req = getPivotExportQuery(stateManagers, true)!;

    try {
      const resp = await $mutation.mutateAsync({
        organization,
        project,
        name: reportName,
        data: {
          options: {
            displayName: values.title,
            refreshCron: refreshCron, // for testing: "* * * * *"
            refreshTimeZone: values.timeZone,
            explore: exploreName,
            queryName: queryName,
            queryArgsJson: JSON.stringify(req.metricsViewAggregationRequest!),
            exportLimit: values.exportLimit || undefined,
            exportIncludeHeader: values.exportIncludeHeader || false,
            exportFormat: values.exportFormat,
            emailRecipients: values.emailRecipients.filter(Boolean),
            slackChannels: values.enableSlackNotification
              ? values.slackChannels.filter(Boolean)
              : undefined,
            slackUsers: values.enableSlackNotification
              ? values.slackUsers.filter(Boolean)
              : undefined,
            webOpenState: get(dashboardStore).proto,
            webOpenMode:
              props.mode === "create"
                ? "recipient" // To be changed to "filtered" once support is added
                : ((props.reportSpec.annotations as V1ReportSpecAnnotations)[
                    "web_open_mode"
                  ] ?? "recipient"), // Backwards compatibility
          },
        },
      });
      const newReportName = props.mode === "create" ? resp.name : reportName;

      if (props.mode === "edit") {
        await queryClient.invalidateQueries({
          queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
            "name.name": reportName,
            "name.kind": ResourceKind.Report,
          }),
        });
      }

      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
      });

      eventBus.emit("notification", {
        message: `Report ${props.mode === "create" ? "created" : "edited"}`,
        type: "success",
      });

      return goto(`/${organization}/${project}/-/reports/${newReportName}`);
    } catch {
      // showing error below
    }
  }

  function handleCancel() {
    if (props.mode === "create") {
      return goto(`/${organization}/${project}/explore/${exploreName}`);
    } else {
      return goto(`/${organization}/${project}/-/reports/${reportName}`);
    }
  }
</script>

<form
  autocomplete="off"
  class="max-h-fit w-[500px] bg-surface flex-none flex flex-col border select-none rounded-[2px]"
  id={FORM_ID}
  on:submit|preventDefault={submit}
  use:enhance
>
  <h1 class="pt-6 px-5">Schedule report</h1>

  <div
    class="px-5 flex flex-col gap-y-3 w-full h-fit overflow-y-auto overflow-x-visible"
  >
    <Input
      bind:value={$form["title"]}
      errors={$errors["title"]}
      id="title"
      label="Report title"
      placeholder="My report"
    />
    <ScheduleForm data={form} {exploreName} />
    <Select
      bind:value={$form["exportFormat"]}
      id="exportFormat"
      label="Format"
      options={[
        { value: V1ExportFormat.EXPORT_FORMAT_CSV, label: "CSV" },
        { value: V1ExportFormat.EXPORT_FORMAT_PARQUET, label: "Parquet" },
        { value: V1ExportFormat.EXPORT_FORMAT_XLSX, label: "XLSX" },
      ]}
    />
    <Input
      bind:value={$form["exportLimit"]}
      errors={$errors["exportLimit"]}
      id="exportLimit"
      label="Row limit"
      optional
      placeholder="1000"
    />
    <div class="flex items-center gap-x-1">
      <Checkbox
        bind:checked={$form["exportIncludeHeader"]}
        id="exportIncludeHeader"
        onCheckedChange={(checked) => {
          $form["exportIncludeHeader"] = Boolean(checked);
        }}
        inverse
        disabled={$form["exportFormat"] ===
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

    <MultiInput
      id="emailRecipients"
      label="Email Recipients"
      hint="Recipients will receive different views based on their security policy.
        Recipients without project access can only download the report."
      bind:values={$form["emailRecipients"]}
      errors={$errors["emailRecipients"]}
      singular="email"
      plural="emails"
      placeholder="Enter an email address"
    />
    {#if $hasSlackNotifier.data}
      <FormSection
        bind:enabled={$form["enableSlackNotification"]}
        showSectionToggle
        title="Slack notifications"
        padding=""
      >
        <MultiInput
          id="slackChannels"
          label="Channels"
          hint="We’ll send alerts directly to these channels."
          bind:values={$form["slackChannels"]}
          errors={$errors["slackChannels"]}
          singular="channel"
          plural="channels"
          placeholder="# Enter a Slack channel name"
        />
        <MultiInput
          id="slackUsers"
          label="Users"
          hint="We’ll alert them with direct messages in Slack."
          bind:values={$form["slackUsers"]}
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

    {#if generalErrors}
      <div class="text-red-500">{generalErrors}</div>
    {/if}
  </div>

  <div class="flex flex-col gap-y-3 mt-auto border-t px-5 pb-6 pt-3">
    <Button onClick={handleCancel}>Cancel</Button>
    <Button
      disabled={$submitting}
      form={FORM_ID}
      submitForm
      type="primary"
      label={props.mode === "create" ? "Create report" : "Save report"}
    >
      {props.mode === "create" ? "Create" : "Save"}
    </Button>
  </div>
</form>
