<script lang="ts">
  import { DialogTitle } from "@rgossiaux/svelte-headlessui";
  import * as DialogTabs from "@rilldata/web-common/components/dialog/tabs";
  import { createEventDispatcher } from "svelte";
  import Button from "../../components/button/Button.svelte";
  import AlertDialogCriteriaTab from "./criteria-tab/AlertDialogCriteriaTab.svelte";
  import AlertDialogDataTab from "./data-tab/AlertDialogDataTab.svelte";
  import AlertDialogDeliveryTab from "./delivery-tab/AlertDialogDeliveryTab.svelte";
  import { checkIsTabValid } from "./form-utils";

  export let formId: string;
  export let formState: any; // svelte-forms-lib's FormState
  export let dialogTitle: string;

  const dispatch = createEventDispatcher();

  const { form, errors, handleSubmit, isSubmitting } = formState;

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

<!-- 602px = 1px border on each side of the form + 3 tabs with a 200px fixed-width -->
<form
  class="transform bg-white rounded-md border border-slate-300 flex flex-col shadow-lg w-[602px]"
  id={formId}
  on:submit|preventDefault={handleSubmit}
>
  <DialogTitle class="px-6 py-4 text-gray-900 text-lg font-semibold leading-7">
    {dialogTitle}
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
        form={formId}
        submitForm
      >
        Create
      </Button>
    {/if}
  </div>
</form>
