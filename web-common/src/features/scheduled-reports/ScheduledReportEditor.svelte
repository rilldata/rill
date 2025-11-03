<script lang="ts" context="module">
  import type { V1ReportSpec } from "@rilldata/web-common/runtime-client";

  export type CreateReportProps = {
    mode: "create";
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
  import { Button } from "@rilldata/web-common/components/button";
  import { getPivotQueryFromExploreState } from "@rilldata/web-common/features/dashboards/pivot/pivot-export.ts";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import BaseScheduledReportForm from "@rilldata/web-common/features/scheduled-reports/BaseScheduledReportForm.svelte";
  import { convertFormValuesToCronExpression } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
  import {
    getExistingReportInitialFormValues,
    getNewReportInitialFormValues,
    type ReportValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { get } from "svelte/store";
  import { defaults, superForm } from "sveltekit-superforms";
  import { type ValidationAdapter, yup } from "sveltekit-superforms/adapters";
  import { array, boolean, object, string } from "yup";

  export let props: CreateReportProps | EditReportProps;
  export let organization: string;
  export let project: string;

  const FORM_ID = "scheduled-report-form";
  const AGGREGATION_QUERY_NAME = "MetricsViewAggregation";
  const MIN_SIDEBAR_WIDTH = 500;
  const MAX_SIDEBAR_WIDTH = 1000;
  const SIDEBAR_WIDTH = 750;
  let width = SIDEBAR_WIDTH;

  $: reportName = props.mode === "create" ? "" : props.reportName;
  let exploreName = "";
  const initialExploreName = props.mode === "create" ? props.exploreName : "";

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewSpec = $exploreSpecQuery.data?.metricsView ?? {};
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: exploreStore = useExploreState(exploreName);

  const user = createAdminServiceGetCurrentUser();

  $: mutation =
    props.mode === "create"
      ? createAdminServiceCreateReport()
      : createAdminServiceEditReport();

  $: initialValues =
    props.mode === "create"
      ? getNewReportInitialFormValues(
          $user.data?.user?.email,
          props.exploreName,
          {},
        )
      : getExistingReportInitialFormValues(
          props.reportSpec,
          $user.data?.user?.email,
          {},
        );

  const schema = yup(
    object({
      title: string().required("Required"),
      exploreName: string().required("Required"),
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

  $: exploreName = $form?.exploreName ?? "";

  $: generalErrors = $errors._errors?.[0] ?? $mutation.error?.message;

  async function handleSubmit(values: ReportValues) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
      values.dayOfMonth,
    );
    const exploreState = get(exploreStore);
    const req = getPivotQueryFromExploreState(
      metricsViewSpec,
      exploreSpec,
      exploreState,
    );

    try {
      await $mutation.mutateAsync({
        org: organization,
        project,
        name: reportName,
        data: {
          options: {
            displayName: values.title,
            refreshCron: refreshCron, // for testing: "* * * * *"
            refreshTimeZone: values.timeZone,
            explore: values.exploreName,
            queryName: AGGREGATION_QUERY_NAME,
            queryArgsJson: JSON.stringify(req.metricsViewAggregationRequest),
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
            webOpenState: exploreState.proto,
            webOpenMode: values.webOpenMode,
          },
        },
      });

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

      if (props.mode === "create") {
        void goto(`/${organization}/${project}/-/reports`);
      } else {
        void goto(`/${organization}/${project}/-/reports/${reportName}`);
      }

      eventBus.emit("notification", {
        message: `Report ${props.mode === "create" ? "created" : "edited"}`,
        type: "success",
      });
    } catch {
      // showing error below
    }
  }

  function handleCancel() {
    if (props.mode === "create" && initialExploreName) {
      return goto(`/${organization}/${project}/explore/${initialExploreName}`);
    } else {
      return goto(`/${organization}/${project}/-/reports/${reportName}`);
    }
  }

  function handleExploreChange(exploreName: string) {
    if (props.mode === "create") {
      void goto(
        `/${organization}/${project}/-/reports/-/create/explore/${exploreName}`,
      );
    } else {
      void goto(
        `/${organization}/${project}/-/reports/${reportName}/edit/explore/${exploreName}`,
      );
    }
  }

  function updateSidebarWidth(newWidth: number): void {
    width = Math.max(MIN_SIDEBAR_WIDTH, Math.min(MAX_SIDEBAR_WIDTH, newWidth));
  }
</script>

<div class="flex flex-col border-l relative h-full" style="width: {width}px;">
  <Resizer
    min={MIN_SIDEBAR_WIDTH}
    max={MAX_SIDEBAR_WIDTH}
    basis={SIDEBAR_WIDTH}
    dimension={width}
    direction="EW"
    side="left"
    onUpdate={updateSidebarWidth}
  />
  <SidebarWrapper title="Scheduled reports">
    <BaseScheduledReportForm
      formId={FORM_ID}
      data={form}
      {errors}
      {submit}
      {enhance}
      height=""
      {handleExploreChange}
    />

    {#if generalErrors}
      <div class="text-red-500">{generalErrors}</div>
    {/if}

    <footer
      class="flex flex-col gap-y-3 mt-auto border-t px-5 pb-6 pt-3"
      slot="footer"
    >
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
    </footer>
  </SidebarWrapper>
</div>
