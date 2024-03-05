<script lang="ts">
  import { page } from "$app/stores";
  import { Dialog, DialogOverlay } from "@rgossiaux/svelte-headlessui";
  import {
    createAdminServiceCreateAlert,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
  import {
    AlertFormValues,
    alertFormValidationSchema,
    getAlertQueryArgsFromFormValues,
  } from "@rilldata/web-common/features/alerts/form-utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    V1Operation,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { get } from "svelte/store";
  import { notifications } from "../../components/notifications";
  import { runtime } from "../../runtime-client/runtime-store";
  import BaseAlertForm from "./BaseAlertForm.svelte";

  export let open: boolean;

  const user = createAdminServiceGetCurrentUser();
  const createAlert = createAdminServiceCreateAlert();
  $: organization = $page.params.organization;
  $: project = $page.params.project;
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const {
    metricsViewName,
    dashboardStore,
    selectors: {
      timeRangeSelectors: { timeControlsState },
    },
  } = getStateManagers();
  const timeControls = get(timeControlsState);

  const formState = createForm({
    initialValues: {
      name: "",
      measure: $dashboardStore.leaderboardMeasureName ?? "",
      splitByDimension: $dashboardStore.selectedDimensionName ?? "",
      splitByTimeGrain: "",
      criteria: [
        {
          field: $dashboardStore.leaderboardMeasureName ?? "",
          operation: V1Operation.OPERATION_GTE,
          value: "0",
        },
      ],
      criteriaOperation: V1Operation.OPERATION_AND,
      snooze: SnoozeOptions[0].value, // Defaults to `Off`
      recipients: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
      // The remaining fields are not editable in the form, but it's helpful to have access to them throughout the alert dialog
      // Also, in the future, they might even be editable.
      metricsViewName: $metricsViewName,
      whereFilter: $dashboardStore.whereFilter,
      timeRange: {
        isoDuration: timeControls.selectedTimeRange?.name,
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
    } as AlertFormValues,
    validationSchema: alertFormValidationSchema,
    onSubmit: async (values) => {
      try {
        await $createAlert.mutateAsync({
          organization,
          project,
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
        queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );
        dispatch("close");
        notifications.send({
          message: "Alert created",
          link: {
            href: `/${organization}/${project}/-/alerts`,
            text: "Go to alerts",
          },
          options: {
            persistedLink: true,
          },
        });
      } catch (e) {
        // showing error below
      }
    },
  });
</script>

<Dialog
  class="fixed inset-0 flex items-center justify-center z-50 overflow-auto"
  {open}
>
  <DialogOverlay
    class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
  />
  <BaseAlertForm {formState} isEditForm={false} on:close />
</Dialog>
