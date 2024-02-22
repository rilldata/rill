<script lang="ts">
  import { page } from "$app/stores";
  import { Dialog, DialogOverlay } from "@rgossiaux/svelte-headlessui";
  import { createAdminServiceEditAlert } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { notifications } from "../../components/notifications";
  import {
    V1AlertSpec,
    V1MetricsViewAggregationRequest,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { SnoozeOptions } from "./delivery-tab/snooze";
  import {
    alertFormValidationSchema,
    getAlertQueryArgsFromFormValues,
    getFormValuesFromAlertQueryArgs,
  } from "./form-utils";

  export let open: boolean;
  export let alertSpec: V1AlertSpec;

  const editAlert = createAdminServiceEditAlert();
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertName = $page.params.alert;
  const queryArgsJson = JSON.parse(
    alertSpec.queryArgsJson as string,
  ) as V1MetricsViewAggregationRequest;
  const metricsView = queryArgsJson?.metricsView ?? "";

  const formState = createForm({
    initialValues: {
      name: alertSpec.title as string,
      ...getFormValuesFromAlertQueryArgs(queryArgsJson),
      snooze: SnoozeOptions[0].value, // TODO: get actual value from alertSpec
      recipients: alertSpec?.emailRecipients?.map((r) => ({ email: r })) ?? [],
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
              intervalDuration: undefined,
              queryName: "MetricsViewAggregation",
              queryArgsJson: JSON.stringify(
                getAlertQueryArgsFromFormValues(values),
              ),
              metricsViewName: metricsView,
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

<Dialog class="fixed inset-0 flex items-center justify-center z-50" {open}>
  <DialogOverlay
    class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
  />
  <BaseAlertForm
    on:close
    formId="edit-alert-form"
    {formState}
    dialogTitle="Edit alert"
  />
</Dialog>
