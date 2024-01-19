<script lang="ts">
  import { TabPanel, TabPanels } from "@rgossiaux/svelte-headlessui";
  import Tab from "@rilldata/web-admin/components/tabs/Tab.svelte";
  import TabGroup from "@rilldata/web-admin/components/tabs/TabGroup.svelte";
  import TabList from "@rilldata/web-admin/components/tabs/TabList.svelte";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { Button } from "../../components/button";
  import Dialog from "../../components/dialog-v2/Dialog.svelte";

  export let open: boolean;

  const dispatch = createEventDispatcher();

  createForm({
    initialValues: {},
    onSubmit: async (values) => {
      console.log("submitting alerts form with these values: ", values);
      dispatch("close");
    },
  });

  // const { isSubmitting, form } = formState;

  const tabs = ["Data", "Criteria", "Delivery"];
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
      <TabPanels>
        <TabPanel>Data tab</TabPanel>
        <TabPanel>Criteria tab</TabPanel>
        <TabPanel>Delivery tab</TabPanel>
      </TabPanels>
    </TabGroup>
  </svelte:fragment>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-5">
      <div class="grow" />
      <Button on:click={() => dispatch("close")} type="secondary">
        Cancel
      </Button>
      <Button form="create-alert-form" submitForm type="primary">Create</Button>
    </div>
  </svelte:fragment>
</Dialog>
