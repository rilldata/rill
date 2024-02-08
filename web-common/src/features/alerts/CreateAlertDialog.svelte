<script lang="ts">
  import {
    Dialog,
    DialogOverlay,
    DialogTitle,
  } from "@rgossiaux/svelte-headlessui";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import AlertDialogCriteriaTab from "web-common/src/features/alerts/criteria-tab/AlertDialogCriteriaTab.svelte";
  import AlertDialogDeliveryTab from "web-common/src/features/alerts/delivery-tab/AlertDialogDeliveryTab.svelte";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import AlertDialogDataTab from "./data-tab/AlertDialogDataTab.svelte";
  import * as DialogTabs from "./dialog-tabs";

  export let open: boolean;

  const dispatch = createEventDispatcher();
  const user = createAdminServiceGetCurrentUser();

  const formState = createForm({
    initialValues: {
      name: "",
      measure: "",
      splitByDimension: "",
      criteria: [
        {
          field: "",
          operation: "",
          value: 0,
        },
      ],
      criteriaOperation: V1Operation.OPERATION_AND,
      snooze: "OFF", // TODO: use enum from backend
      recipients: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
    },
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
      console.log("submitting alerts form with these values: ", values);
      // dispatch("close");
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
  $: isTabValid = checkIsTabValid(selectedTabIndex, $form, $errors);

  function checkIsTabValid(
    tabIndex: number,
    form: Record<string, any>,
    errors: Record<string, string>,
  ): boolean {
    let hasRequiredFields: boolean;
    let hasErrors: boolean;

    if (tabIndex === 0) {
      hasRequiredFields = form.name !== "" && form.measure !== "";
      hasErrors = !!errors.name && !!errors.measure;
    } else if (tabIndex === 1) {
      hasRequiredFields = true;
      form.criteria.forEach((criteria) => {
        if (
          criteria.field === "" ||
          criteria.operation === "" ||
          criteria.value === ""
        ) {
          hasRequiredFields = false;
        }
      });
      hasErrors = false;
      (errors.criteria as unknown as any[]).forEach((criteriaError) => {
        if (
          criteriaError.field ||
          criteriaError.operation ||
          criteriaError.value
        ) {
          hasErrors = true;
        }
      });
    } else if (tabIndex === 2) {
      // TODO: do better for >1 recipients
      hasRequiredFields = form.snooze !== "" && form.recipients[0].email !== "";
      hasErrors = !!errors.snooze || !!errors.recipients[0].email;
    } else {
      throw new Error(`Unexpected tabIndex: ${tabIndex}`);
    }

    return hasRequiredFields && !hasErrors;
  }

  let selectedTabIndex = 0;

  function handleCancel() {
    dispatch("close");
  }

  function handleBack() {
    selectedTabIndex -= 1;
  }

  function handleNextTab() {
    selectedTabIndex += 1;
  }
</script>

<Dialog {open} class="fixed inset-0 flex items-center justify-center z-50">
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
    <DialogTabs.Root value={tabs[selectedTabIndex]}>
      <DialogTabs.List class="border-t border-gray-200">
        {#each tabs as tab, i}
          <DialogTabs.Trigger value={tab} tabIndex={i}>
            {tab}
          </DialogTabs.Trigger>
        {/each}
      </DialogTabs.List>
      <div class="p-3 bg-slate-100">
        <DialogTabs.Content value={tabs[0]}>
          <AlertDialogDataTab {formState} />
        </DialogTabs.Content>
        <DialogTabs.Content value={tabs[1]}>
          <AlertDialogCriteriaTab {formState} />
        </DialogTabs.Content>
        <DialogTabs.Content value={tabs[2]}>
          <AlertDialogDeliveryTab {formState} />
        </DialogTabs.Content>
      </div>
    </DialogTabs.Root>
    <div class="px-6 py-3 flex items-center gap-x-2">
      <div class="grow" />
      {#if selectedTabIndex === 0}
        <Button on:click={handleCancel} type="secondary">Cancel</Button>
      {:else}
        <Button on:click={handleBack} type="secondary">Back</Button>
      {/if}
      {#if selectedTabIndex !== 2}
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
