<script lang="ts">
  import { page } from "$app/stores";
  import { Dialog, DialogOverlay } from "@rgossiaux/svelte-headlessui";
  import { createAdminServiceEditAlert } from "@rilldata/web-admin/client";
  import { extractAlertFormValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { notifications } from "../../components/notifications";
  import {
    V1AlertSpec,
    V1MetricsViewAggregationRequest,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
    type V1MetricsViewSpec,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { getSnoozeValueFromAlertSpec } from "./delivery-tab/snooze";
  import {
    alertFormValidationSchema,
    getAlertQueryArgsFromFormValues,
  } from "./form-utils";

  export let open: boolean;
  export let alertSpec: V1AlertSpec;
  // Since form state is not loaded async we need to make sure this is fetched before this component
  // So it is easier to get it from the parent
  export let metricsViewSpec: V1MetricsViewSpec;

  const editAlert = createAdminServiceEditAlert();
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertName = $page.params.alert;
  const queryArgsJson = JSON.parse(
    alertSpec.queryArgsJson as string,
  ) as V1MetricsViewAggregationRequest;

  const formState = createForm({
    initialValues: {
      name: alertSpec.title as string,
      snooze: getSnoozeValueFromAlertSpec(alertSpec),
      recipients: alertSpec?.emailRecipients?.map((r) => ({ email: r })) ?? [],
      splitByTimeGrain: alertSpec.intervalsIsoDuration ?? "",
      ...extractAlertFormValues(queryArgsJson, metricsViewSpec),
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
              title: values.name,
              intervalDuration: values.splitByTimeGrain,
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
