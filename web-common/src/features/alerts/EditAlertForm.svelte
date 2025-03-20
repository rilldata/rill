<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceEditAlert } from "@rilldata/web-admin/client";
  import { useAlertDashboardState } from "@rilldata/web-admin/features/alerts/selectors";
  import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
  import {
    extractAlertFormValues,
    extractAlertNotification,
  } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
  import {
    useMetricsViewTimeRange,
    useMetricsViewValidSpec,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import {
    type V1AlertSpec,
    type V1MetricsViewAggregationRequest,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { getSnoozeValueFromAlertSpec } from "./delivery-tab/snooze";
  import {
    alertFormValidationSchema,
    type AlertFormValues,
    getAlertQueryArgsFromFormValues,
  } from "./form-utils";

  export let alertSpec: V1AlertSpec;
  export let metricsViewName: string;

  const editAlert = createAdminServiceEditAlert();
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  $: ({ instanceId } = $runtime);

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertName = $page.params.alert;
  const queryArgsJson = JSON.parse(
    (alertSpec.resolverProperties?.query_args_json ??
      alertSpec.queryArgsJson) as string,
  ) as V1MetricsViewAggregationRequest;

  $: metricsViewSpec = useMetricsViewValidSpec(instanceId, metricsViewName);
  $: timeRange = useMetricsViewTimeRange(instanceId, metricsViewName, {
    query: { queryClient },
  });

  const exploreName = getExploreName(
    alertSpec.annotations?.web_open_path ?? "",
  );
  const webState = alertSpec.annotations?.web_open_state ?? "";
  $: dashboardState = useAlertDashboardState(instanceId, alertSpec);

  const formState = createForm<AlertFormValues>({
    initialValues: {
      name: alertSpec.displayName as string,
      exploreName: exploreName ?? metricsViewName,
      snooze: getSnoozeValueFromAlertSpec(alertSpec),
      evaluationInterval: alertSpec.intervalsIsoDuration ?? "",
      ...extractAlertNotification(alertSpec),
      ...extractAlertFormValues(
        queryArgsJson,
        $metricsViewSpec?.data ?? {},
        $timeRange?.data ?? {},
        $dashboardState.data ?? {},
      ),
    },
    validationSchema: alertFormValidationSchema,
    onSubmit: async (values) => {
      try {
        await $editAlert.mutateAsync({
          organization,
          project,
          name: alertName,
          data: {
            options: {
              displayName: values.name,
              queryName: "MetricsViewAggregation",
              queryArgsJson: JSON.stringify(
                getAlertQueryArgsFromFormValues(values),
              ),
              metricsViewName: values.metricsViewName,
              slackChannels: values.enableSlackNotification
                ? values.slackChannels.map((c) => c.channel).filter(Boolean)
                : undefined,
              slackUsers: values.enableSlackNotification
                ? values.slackUsers.map((c) => c.email).filter(Boolean)
                : undefined,
              emailRecipients: values.enableEmailNotification
                ? values.emailRecipients.map((r) => r.email).filter(Boolean)
                : undefined,
              renotify: !!values.snooze,
              renotifyAfterSeconds: values.snooze ? Number(values.snooze) : 0,
              webOpenPath: exploreName ? `/explore/${exploreName}` : undefined,
              // TODO: if we ever allow users to update the dashboard filters in edit then we need to update this
              //       it would involve getting fields from "values" and converting to proto
              webOpenState: webState,
            },
          },
        });
        void queryClient.invalidateQueries(
          getRuntimeServiceGetResourceQueryKey(instanceId, {
            "name.name": alertName,
            "name.kind": ResourceKind.Alert,
          }),
        );
        void queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey(instanceId),
        );
        dispatch("close");
        eventBus.emit("notification", {
          message: "Alert edited",
          type: "success",
        });
      } catch {
        // showing error below
      }
    },
  });
  const { form } = formState;
  $: if ($metricsViewSpec?.data && $timeRange?.data) {
    const formValues = extractAlertFormValues(
      queryArgsJson,
      $metricsViewSpec.data,
      $timeRange.data,
      dashboardState,
    );
    for (const fk in formValues) {
      $form[fk] = formValues[fk];
    }
  }
</script>

<BaseAlertForm {formState} isEditForm={true} on:cancel on:close />
