<script lang="ts">
  import { page } from "$app/stores";
  import {
    Dialog,
    DialogOverlay,
    DialogTitle,
  } from "@rgossiaux/svelte-headlessui";
  import {
    createAdminServiceCreateAlert,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import * as DialogTabs from "@rilldata/web-common/components/dialog/tabs";
  import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
  import {
    type AlertFormValue,
    checkIsTabValid,
    getAlertQueryArgs,
  } from "@rilldata/web-common/features/alerts/form-utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    V1Operation,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import { notifications } from "../../components/notifications";
  import { runtime } from "../../runtime-client/runtime-store";
  import AlertDialogCriteriaTab from "./criteria-tab/AlertDialogCriteriaTab.svelte";
  import AlertDialogDataTab from "./data-tab/AlertDialogDataTab.svelte";
  import AlertDialogDeliveryTab from "./delivery-tab/AlertDialogDeliveryTab.svelte";

  export let open: boolean;

  const user = createAdminServiceGetCurrentUser();
  const createAlert = createAdminServiceCreateAlert();
  $: organization = $page.params.organization;
  $: project = $page.params.project;
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const { metricsViewName, dashboardStore } = getStateManagers();

  const formState = createForm({
    initialValues: {
      name: "",
      measure: "",
      splitByDimension: "",
      criteria: [
        {
          field: "",
          operation: "",
          value: "0",
        },
      ],
      criteriaOperation: V1Operation.OPERATION_AND,
      snooze: SnoozeOptions[0].value,
      recipients: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
    } as AlertFormValue,
    validationSchema: yup.object({
      name: yup.string().required("Required"),
      measure: yup.string().required("Required"),
      criteria: yup.array().of(
        yup.object().shape({
          field: yup.string().required("Required"),
          operation: yup.string().required("Required"),
          value: yup.number().required("Required"),
        }),
      ),
      criteriaOperation: yup.string().required("Required"),
      snooze: yup.string().required("Required"),
      recipients: yup.array().of(
        yup.object().shape({
          email: yup.string().email("Invalid email"),
        }),
      ),
    }),
    onSubmit: async (values) => {
      try {
        await $createAlert.mutateAsync({
          organization,
          project,
          data: {
            options: {
              title: values.name,
              intervalDuration: undefined, // TODO: this is the "for every" field I think?
              queryName: "MetricsViewAggregation",
              queryArgsJson: JSON.stringify(
                getAlertQueryArgs($metricsViewName, values, $dashboardStore),
              ),
              metricsViewName: $metricsViewName,
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

  const { form, isSubmitting, errors, handleSubmit } = formState;

  const tabs = ["Data", "Criteria", "Delivery"];

  /**
   * Because this form's fields are spread over multiple tabs, we implement our own `isValid` logic for each tab.
   * A tab is valid (i.e. it's okay to proceed to the next tab) if:
   * 1) The tab's required fields are filled out
   * 2) The tab's fields don't have errors.
   */
  $: isTabValid = checkIsTabValid(currentTabIndex, $form, $errors);

  let currentTabIndex = 0;

  function handleCancel() {
    dispatch("close");
  }

  function handleBack() {
    currentTabIndex -= 1;
  }

  function handleNextTab() {
    currentTabIndex += 1;
  }
</script>

<Dialog class="fixed inset-0 flex items-center justify-center z-50" {open}>
  <DialogOverlay
    class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
  />
  <!-- 602px = 1px border on each side of the form + 3 tabs with a 200px fixed-width -->
  <form
    class="transform bg-white rounded-md border border-slate-300 flex flex-col shadow-lg w-[602px]"
    id="create-alert-form"
    on:submit|preventDefault={handleSubmit}
  >
    <DialogTitle
      class="px-6 py-4 text-gray-900 text-lg font-semibold leading-7"
    >
      Create alert
    </DialogTitle>
    <DialogTabs.Root value={tabs[currentTabIndex]}>
      <DialogTabs.List class="border-t border-gray-200">
        {#each tabs as tab, i}
          <DialogTabs.Trigger value={tab} tabIndex={i}>
            {tab}
          </DialogTabs.Trigger>
        {/each}
      </DialogTabs.List>
      <div class="p-3 bg-slate-100">
        <DialogTabs.Content {currentTabIndex} tabIndex={0} value={tabs[0]}>
          <AlertDialogDataTab {formState} />
        </DialogTabs.Content>
        <DialogTabs.Content {currentTabIndex} tabIndex={1} value={tabs[1]}>
          <AlertDialogCriteriaTab {formState} />
        </DialogTabs.Content>
        <DialogTabs.Content {currentTabIndex} tabIndex={2} value={tabs[2]}>
          <AlertDialogDeliveryTab {formState} />
        </DialogTabs.Content>
      </div>
    </DialogTabs.Root>
    <div class="px-6 py-3 flex items-center gap-x-2">
      <div class="grow" />
      {#if currentTabIndex === 0}
        <Button on:click={handleCancel} type="secondary">Cancel</Button>
      {:else}
        <Button on:click={handleBack} type="secondary">Back</Button>
      {/if}
      {#if currentTabIndex !== 2}
        <Button type="primary" disabled={!isTabValid} on:click={handleNextTab}>
          Next
        </Button>
      {:else}
        <Button
          type="primary"
          disabled={!isTabValid || $isSubmitting}
          form="create-alert-form"
          submitForm
        >
          Create
        </Button>
      {/if}
    </div>
  </form>
</Dialog>
