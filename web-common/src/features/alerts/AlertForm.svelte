<script lang="ts" context="module">
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";

  export type CreateAlertProps = {
    mode: "create";
    exploreName: string;
  };

  export type EditAlertProps = {
    mode: "edit";
    alertSpec: V1AlertSpec;
  };
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateAlert,
    createAdminServiceEditAlert,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import {
    getAlertDashboardName,
    unwrapQueryData,
    useAlertDashboardState,
  } from "@rilldata/web-admin/features/alerts/selectors.ts";
  import { DialogTitle } from "@rilldata/web-common/components/dialog";
  import * as DialogTabs from "@rilldata/web-common/components/dialog/tabs";
  import {
    getNewAlertInitialFiltersFormValues,
    getNewAlertInitialFormValues,
  } from "@rilldata/web-common/features/alerts/create-alert-utils.ts";
  import AlertDialogCriteriaTab from "@rilldata/web-common/features/alerts/criteria-tab/AlertDialogCriteriaTab.svelte";
  import AlertDialogDataTab from "@rilldata/web-common/features/alerts/data-tab/AlertDialogDataTab.svelte";
  import AlertDialogDeliveryTab from "@rilldata/web-common/features/alerts/delivery-tab/AlertDialogDeliveryTab.svelte";
  import { getExistingAlertInitialFormValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values.ts";
  import {
    alertFormValidationSchema,
    type AlertFormValues,
    checkIsTabValid,
    FieldsByTab,
    getAlertQueryArgsFromFormValues,
  } from "@rilldata/web-common/features/alerts/form-utils.ts";
  import {
    generateAlertName,
    isSomeFieldTainted,
  } from "@rilldata/web-common/features/alerts/utils.ts";
  import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto.ts";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { convertFormValuesToCronExpression } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
  import { getFiltersAndTimeControlsFromAggregationRequest } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { X } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import Button from "web-common/src/components/button/Button.svelte";

  export let onClose: () => void;
  export let onCancel: () => void;
  export let props: CreateAlertProps | EditAlertProps;

  const user = createAdminServiceGetCurrentUser();

  $: ({ organization, project, alert: alertName } = $page.params);
  $: ({ instanceId } = $runtime);

  // Convenience variable to be used when other fields from props are not needed.
  // Typescript won't parse the object switch if this is used in conditionals, so some statements below don't use this.
  $: isCreateForm = props.mode === "create";

  $: exploreName =
    props.mode === "create"
      ? props.exploreName
      : getAlertDashboardName(props.alertSpec);

  $: validExploreSpec = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewSpec = $validExploreSpec.data?.metricsView ?? {};
  $: exploreSpec = $validExploreSpec.data?.explore ?? {};
  $: metricsViewName = exploreSpec.metricsView ?? "";

  $: allTimeRangeResp = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    undefined,
    queryClient,
  );

  $: exploreState =
    props.mode === "create"
      ? useExploreState(props.exploreName)
      : unwrapQueryData(useAlertDashboardState(instanceId, props.alertSpec));

  $: mutation =
    props.mode === "create"
      ? createAdminServiceCreateAlert()
      : createAdminServiceEditAlert();

  $: initialValues =
    props.mode === "create"
      ? getNewAlertInitialFormValues(
          metricsViewName,
          exploreName,
          $exploreState!,
          $user.data?.user,
        )
      : getExistingAlertInitialFormValues(props.alertSpec, metricsViewName);

  $: ({ filters, timeControls } =
    props.mode === "create"
      ? getNewAlertInitialFiltersFormValues(
          instanceId,
          metricsViewName,
          exploreName,
          $exploreState!,
        )
      : getFiltersAndTimeControlsFromAggregationRequest(
          instanceId,
          metricsViewName,
          exploreName,
          JSON.parse(
            props.alertSpec.queryArgsJson ||
              (props.alertSpec.resolverProperties?.query_args_json as
                | string
                | undefined) ||
              "{}",
          ),
          $allTimeRangeResp.data?.timeRangeSummary,
        ));
  $: ({ selectedComparisonTimeRange } = timeControls);

  $: superFormInstance = superForm(
    defaults(initialValues, alertFormValidationSchema),
    {
      SPA: true,
      validators: alertFormValidationSchema,
      dataType: "json",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        return handleSubmit(values);
      },
      validationMethod: "oninput",
      invalidateAll: false,
    },
  );
  $: ({ form, errors, enhance, submit, submitting, tainted, validate } =
    superFormInstance);

  $: formId = isCreateForm ? "create-alert-form" : "edit-alert-form";
  $: dialogTitle = isCreateForm ? "Create Alert" : "Edit Alert";

  const tabs = ["Data", "Criteria", "Delivery"];

  /**
   * Because this form's fields are spread over multiple tabs, we implement our own `isValid` logic for each tab.
   * A tab is valid (i.e. it's okay to proceed to the next tab) if:
   * 1) The tab's required fields are filled out
   * 2) The tab's fields don't have errors.
   */
  $: isTabValid = checkIsTabValid(currentTabIndex, $form, $errors);

  let currentTabIndex = 0;

  async function handleSubmit(values: AlertFormValues) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
      values.dayOfMonth,
    );

    await $mutation.mutateAsync({
      org: organization,
      project,
      name: alertName,
      data: {
        options: {
          displayName: values.name,
          queryName: "MetricsViewAggregation",
          queryArgsJson: JSON.stringify(
            getAlertQueryArgsFromFormValues(
              values,
              filters.toState(),
              timeControls.toState(),
              exploreSpec,
            ),
          ),
          metricsViewName: values.metricsViewName,
          slackChannels: values.enableSlackNotification
            ? values.slackChannels.filter(Boolean)
            : undefined,
          slackUsers: values.enableSlackNotification
            ? values.slackUsers.filter(Boolean)
            : undefined,
          emailRecipients: values.enableEmailNotification
            ? values.emailRecipients.filter(Boolean)
            : undefined,
          refreshCron: !values.refreshWhenDataRefreshes ? refreshCron : "", // for testing: "* * * * *"
          refreshTimeZone: values.timeZone,
          renotify: !!values.snooze,
          renotifyAfterSeconds: values.snooze ? Number(values.snooze) : 0,
          webOpenPath: `/explore/${encodeURIComponent(exploreName)}`,
          webOpenState: getProtoFromDashboardState(
            $exploreState as ExploreState,
            exploreSpec,
          ),
        },
      },
    });
    if (!isCreateForm) {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
          "name.name": alertName,
          "name.kind": ResourceKind.Alert,
        }),
      });
    }
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
    });
    onClose();
    if (isCreateForm) {
      eventBus.emit("notification", {
        message: "Alert created",
        link: {
          href: `/${organization}/${project}/-/alerts`,
          text: "Go to alerts",
        },
      });
    } else {
      eventBus.emit("notification", {
        message: "Alert edited",
        type: "success",
      });
    }
  }

  $: measure = $form.measure;
  function measureUpdated(mes: string) {
    $form.criteria.forEach((c) => (c.measure = mes));
    $form.criteria = [...$form.criteria];
  }
  $: measureUpdated(measure);

  function handleCancel() {
    if ($tainted && isSomeFieldTainted($tainted)) {
      onCancel();
    } else {
      onClose();
    }
  }

  function handleBack() {
    currentTabIndex -= 1;
  }

  function handleNextTab() {
    if (!isTabValid) {
      FieldsByTab[currentTabIndex]?.forEach((field) => validate(field as any));
      return;
    }
    currentTabIndex += 1;

    if (!isCreateForm || currentTabIndex !== 2 || $tainted?.name) {
      return;
    }
    // if the user came to the delivery tab and name was not changed then auto generate it
    const name = generateAlertName(
      $form,
      $selectedComparisonTimeRange,
      metricsViewSpec,
    );
    if (!name) return;
    $form.name = name;
  }
</script>

<form
  autocomplete="off"
  class="flex flex-col gap-y-3"
  id={formId}
  on:submit|preventDefault={submit}
  use:enhance
>
  <DialogTitle
    class="px-6 py-4 text-gray-900 text-lg font-semibold leading-7 flex flex-row items-center justify-between"
  >
    <div>{dialogTitle}</div>
    <Button type="link" noStroke compact onClick={handleCancel}>
      <X strokeWidth={3} size={16} class="text-black" />
    </Button>
  </DialogTitle>
  <DialogTabs.Root value={tabs[currentTabIndex]}>
    <DialogTabs.List class="border-t">
      {#each tabs as tab, i (i)}
        <!-- inner width is 800px. so, width = ceil(800/3) = 267 -->
        <DialogTabs.Trigger value={tab} tabIndex={i} class="w-[267px]">
          {tab}
        </DialogTabs.Trigger>
      {/each}
    </DialogTabs.List>
    <div class="p-3 bg-slate-100 h-[600px] overflow-auto">
      <DialogTabs.Content {currentTabIndex} tabIndex={0} value={tabs[0]}>
        <AlertDialogDataTab {superFormInstance} {filters} {timeControls} />
      </DialogTabs.Content>
      <DialogTabs.Content {currentTabIndex} tabIndex={1} value={tabs[1]}>
        <AlertDialogCriteriaTab {superFormInstance} {filters} {timeControls} />
      </DialogTabs.Content>
      <DialogTabs.Content {currentTabIndex} tabIndex={2} value={tabs[2]}>
        <AlertDialogDeliveryTab {superFormInstance} {exploreName} />
      </DialogTabs.Content>
    </div>
  </DialogTabs.Root>
  <div class="px-6 py-3 flex items-center gap-x-2">
    <div class="grow" />
    {#if currentTabIndex === 0}
      <Button onClick={handleCancel} type="secondary">Cancel</Button>
    {:else}
      <Button onClick={handleBack} type="secondary">Back</Button>
    {/if}
    {#if currentTabIndex !== 2}
      <Button type="primary" onClick={handleNextTab}>Next</Button>
    {:else}
      <Button type="primary" disabled={$submitting} form={formId} submitForm>
        {isCreateForm ? "Create" : "Update"}
      </Button>
    {/if}
  </div>
</form>
