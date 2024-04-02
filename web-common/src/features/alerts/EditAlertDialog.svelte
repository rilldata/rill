<script lang="ts">
  import { page } from "$app/stores";
  import { Dialog, DialogOverlay } from "@rgossiaux/svelte-headlessui";
  import { createAdminServiceEditAlert } from "@rilldata/web-admin/client";
  import {
    extractAlertFormValueFromComparison,
    extractAlertFormValues,
  } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
  import {
    useMetricsView,
    useMetricsViewTimeRange,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { notifications } from "../../components/notifications";
  import {
    V1AlertSpec,
    V1MetricsViewAggregationRequest,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
    type V1MetricsViewComparisonRequest,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { getSnoozeValueFromAlertSpec } from "./delivery-tab/snooze";
  import {
    alertFormValidationSchema,
    AlertFormValues,
    getAlertQueryArgsFromFormValues,
  } from "./form-utils";

  export let open: boolean;
  export let alertSpec: V1AlertSpec;
  export let metricsViewName: string;

  const editAlert = createAdminServiceEditAlert();
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertName = $page.params.alert;
  const queryArgsJson = JSON.parse(alertSpec.queryArgsJson as string) as
    | V1MetricsViewAggregationRequest
    | V1MetricsViewComparisonRequest;

  $: metricsViewSpec = useMetricsView($runtime?.instanceId, metricsViewName);
  $: timeRange = useMetricsViewTimeRange(
    $runtime?.instanceId,
    metricsViewName,
    { query: { queryClient } },
  );

  const formState = createForm({
    initialValues: {
      name: alertSpec.title as string,
      snooze: getSnoozeValueFromAlertSpec(alertSpec),
      recipients: alertSpec?.emailRecipients?.map((r) => ({ email: r })) ?? [],
      evaluationInterval: alertSpec.intervalsIsoDuration ?? "",
      ...("metricsView" in queryArgsJson
        ? extractAlertFormValues(
            queryArgsJson,
            $metricsViewSpec?.data ?? {},
            $timeRange?.data ?? {},
          )
        : extractAlertFormValueFromComparison(
            queryArgsJson,
            $metricsViewSpec?.data ?? {},
            $timeRange?.data ?? {},
          )),
    } as AlertFormValues,
    validationSchema: alertFormValidationSchema,
    onSubmit: async (values) => {
      try {
        await $editAlert.mutateAsync({
          organization,
          project,
          name: alertName,
          data: {
            options: {
              title: values.name,
              queryName: "MetricsViewAggregation",
              queryArgsJson: JSON.stringify(
                getAlertQueryArgsFromFormValues(values),
              ),
              metricsViewName: values.metricsViewName,
              recipients: values.recipients.map((r) => r.email).filter(Boolean),
              emailRenotify: !!values.snooze,
              emailRenotifyAfterSeconds: values.snooze
                ? Number(values.snooze)
                : 0,
            },
          },
        });
        void queryClient.invalidateQueries(
          getRuntimeServiceGetResourceQueryKey($runtime.instanceId, {
            "name.name": alertName,
            "name.kind": ResourceKind.Alert,
          }),
        );
        void queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );
        dispatch("close");
        notifications.send({
          message: "Alert edited",
          type: "success",
        });
      } catch (e) {
        // showing error below
      }
    },
  });
  const { form } = formState;
  $: if ($metricsViewSpec?.data && $timeRange?.data) {
    let formValues: Record<string, any>;
    if ("metricsView" in queryArgsJson) {
      formValues = extractAlertFormValues(
        queryArgsJson,
        $metricsViewSpec.data,
        $timeRange.data,
      );
    } else {
      formValues = extractAlertFormValueFromComparison(
        queryArgsJson,
        $metricsViewSpec.data,
        $timeRange.data,
      );
    }
    for (const fk in formValues) {
      $form[fk] = formValues[fk];
    }
  }
</script>

<Dialog
  class="fixed inset-0 flex items-center justify-center overflow-auto z-50"
  {open}
>
  <DialogOverlay
    class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
  />
  <BaseAlertForm {formState} isEditForm={true} on:close />
</Dialog>
