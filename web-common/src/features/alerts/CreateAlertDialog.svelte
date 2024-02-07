<script lang="ts">
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import AlertDialogCriteriaTab from "web-common/src/features/alerts/criteria-tab/AlertDialogCriteriaTab.svelte";
  import AlertDialogDeliveryTab from "web-common/src/features/alerts/delivery-tab/AlertDialogDeliveryTab.svelte";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import Dialog from "../../components/dialog-v2/Dialog.svelte";
  import * as Tabs from "../../components/tabs";
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
      dispatch("close");
    },
  });

  const { form, isSubmitting, errors } = formState;

  const tabs = ["Data", "Criteria", "Delivery"];

  /**
   * Because this form's fields are spread over multiple tabs, we implement our own `isValid` logic for each tab.
   * A tab is valid (i.e. it's okay to proceed to the next tab) if:
   * 1) The tab's required fields are filled out
   * 2) The tab's fields don't have errors.
   */
  $: isTabValid = $form && checkIsTabValid(selectedTabIndex, $errors);

  function checkIsTabValid(
    tabIndex: number,
    errors: Record<string, string>,
  ): boolean {
    let hasRequiredFields: boolean;
    let hasErrors: boolean;

    if (tabIndex === 0) {
      hasRequiredFields = $form.name !== "" && $form.measure !== "";
      hasErrors = !!errors.name && !!errors.measure;
    } else if (tabIndex === 1) {
      // TODO
      hasRequiredFields = false;
      hasErrors = true;
    } else if (tabIndex === 2) {
      // TODO
      hasRequiredFields = false;
      hasErrors = true;
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

<Dialog {open} titleMarginBottomOverride="mb-1" widthOverride="600px">
  <svelte:fragment slot="title">Create alert</svelte:fragment>
  <div class="overflow-auto" slot="body">
    <DialogTabs.Root value={tabs[selectedTabIndex]}>
      <DialogTabs.List>
        {#each tabs as tab, i}
          <DialogTabs.Trigger value={tab} tabIndex={i}>
            {tab}
          </DialogTabs.Trigger>
        {/each}
      </DialogTabs.List>
      <Tabs.Content value={tabs[0]}>
        <AlertDialogDataTab {formState} />
      </Tabs.Content>
      <Tabs.Content value={tabs[1]}>
        <AlertDialogCriteriaTab {formState} />
      </Tabs.Content>
      <Tabs.Content value={tabs[2]}>
        <AlertDialogDeliveryTab {formState} />
      </Tabs.Content>
      <div class="flex items-center gap-x-2 mt-5">
        <div class="grow" />
        <Button
          on:click={selectedTabIndex === 0 ? handleCancel : handleBack}
          type="secondary"
        >
          {selectedTabIndex === 0 ? "Cancel" : "Back"}
        </Button>
        <Button
          on:click={selectedTabIndex === 2 ? undefined : handleNextTab}
          disabled={!isTabValid || $isSubmitting}
          form={selectedTabIndex === 2 ? "create-alert-form" : undefined}
          submitForm={selectedTabIndex === 2}
          type="primary"
        >
          {selectedTabIndex === 2 ? "Create" : "Next"}
        </Button>
      </div>
    </DialogTabs.Root>
  </div>
</Dialog>
