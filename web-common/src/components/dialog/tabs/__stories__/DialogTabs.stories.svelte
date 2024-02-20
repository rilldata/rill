<script lang="ts">
  import {
    Dialog,
    DialogOverlay,
    DialogTitle,
  } from "@rgossiaux/svelte-headlessui";
  import { Meta, Story } from "@storybook/addon-svelte-csf";
  import * as DialogTabs from "../";
  import Button from "../../../button/Button.svelte";

  let currentTabIndex = 0;
  const tabs = ["Tab 1", "Tab 2", "Tab 3"];

  function handleNextTab() {
    currentTabIndex += 1;
  }

  function handleBack() {
    currentTabIndex -= 1;
  }

  function handleCancel() {
    console.log("Cancel");
  }

  function handleDone() {
    console.log("Done");
  }
</script>

<Meta title="Components/DialogTabs" />

<Story name="Multi-panel dialog">
  <Dialog
    open={true}
    class="fixed inset-0 flex items-center justify-center z-50"
  >
    <DialogOverlay
      class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
    />
    <!-- 602px = 1px border on each side of the form + 3 tabs with a 200px fixed-width -->
    <form
      class="transform bg-white rounded-md border border-slate-300 flex flex-col shadow-lg w-[602px]"
      id="create-alert-form"
    >
      <DialogTitle
        class="px-6 py-4 text-gray-900 text-lg font-semibold leading-7"
      >
        Multi-panel dialog
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
          <DialogTabs.Content value={tabs[0]} tabIndex={0} {currentTabIndex}>
            <section class="h-56 bg-red-200"></section>
          </DialogTabs.Content>
          <DialogTabs.Content value={tabs[1]} tabIndex={1} {currentTabIndex}>
            <section class="h-40 bg-green-200"></section>
          </DialogTabs.Content>
          <DialogTabs.Content value={tabs[2]} tabIndex={2} {currentTabIndex}>
            <section class="h-64 bg-blue-200"></section>
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
          <Button type="primary" on:click={handleNextTab}>Next</Button>
        {:else}
          <Button type="primary" on:click={handleDone}>Done</Button>
        {/if}
      </div>
    </form>
  </Dialog>
</Story>
