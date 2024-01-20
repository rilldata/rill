<script lang="ts">
  import { TabPanel, TabPanels } from "@rgossiaux/svelte-headlessui";
  import Tab from "@rilldata/web-admin/components/tabs/Tab.svelte";
  import TabGroup from "@rilldata/web-admin/components/tabs/TabGroup.svelte";
  import TabList from "@rilldata/web-admin/components/tabs/TabList.svelte";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import Dialog from "../../components/dialog-v2/Dialog.svelte";
  import AlertDialogCriteriaTab from "./AlertDialogCriteriaTab.svelte";
  import AlertDialogDataTab from "./AlertDialogDataTab.svelte";
  import AlertDialogDeliveryTab from "./AlertDialogDeliveryTab.svelte";

  export let open: boolean;

  const dispatch = createEventDispatcher();

  const formState = createForm({
    initialValues: {
      name: "",
    },
    validationSchema: yup.object({
      name: yup.string().required("Required"),
    }),
    onSubmit: async (values) => {
      console.log("submitting alerts form with these values: ", values);
      dispatch("close");
    },
  });

  const { isSubmitting, errors, touched } = formState;

  const tabs = ["Data", "Criteria", "Delivery"];

  /**
   * Because this form's fields are spread over multiple tabs, we implement our own `isValid` logic for each tab.
   * A tab is valid (i.e. it's okay to proceed to the next tab) if:
   * 1) The tab's required fields have been touched
   * 2) The tab's fields don't have errors.
   */
  function isTabValid(
    tabIndex: number,
    touched: Record<string, boolean>,
    errors: Record<string, string>,
  ): boolean {
    let tabTouched: boolean;
    let tabErrors: boolean;
    if (tabIndex === 0) {
      tabTouched = touched.name;
      tabErrors = !!errors.name;
    } else if (tabIndex === 1) {
      // TODO
      tabTouched = false;
      tabErrors = true;
    } else if (tabIndex === 2) {
      // TODO
      tabTouched = false;
      tabErrors = true;
    } else {
      throw new Error(`Unexpected tabIndex: ${tabIndex}`);
    }

    return tabTouched && !tabErrors;
  }
</script>

<Dialog {open} titleMarginBottomOverride="mb-1">
  <svelte:fragment slot="title">Create alert</svelte:fragment>
  <svelte:fragment slot="body">
    <!-- TODO: match Figma mocks -->
    <!-- TODO: tabs shouldn't be clickable -->
    <TabGroup>
      <TabList>
        {#each tabs as tab}
          <Tab>
            {tab}
          </Tab>
        {/each}
      </TabList>
      <TabPanels let:selectedIndex={selectedTabIndex}>
        <TabPanel>
          <AlertDialogDataTab {formState} />
        </TabPanel>
        <TabPanel>
          <AlertDialogCriteriaTab {formState} />
        </TabPanel>
        <TabPanel>
          <AlertDialogDeliveryTab {formState} />
        </TabPanel>
        <div class="flex items-center gap-x-2 mt-5">
          <div class="grow" />
          <Button on:click={() => dispatch("close")} type="secondary">
            Cancel
          </Button>
          {#if selectedTabIndex !== null}
            <Button
              disabled={!isTabValid(selectedTabIndex, $touched, $errors) ||
                $isSubmitting}
              form={selectedTabIndex === 2 ? "create-alert-form" : undefined}
              submitForm={selectedTabIndex === 2}
              type="primary"
            >
              {selectedTabIndex === 2 ? "Create" : "Next"}
            </Button>
          {/if}
        </div>
      </TabPanels>
    </TabGroup>
  </svelte:fragment>
</Dialog>
