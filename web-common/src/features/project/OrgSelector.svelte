<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    SelectSeparator,
    SelectItem,
  } from "@rilldata/web-common/components/select";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let open = false;
  export let orgs: string[];
  export let onSelect: (org: string) => void;

  let selectedOrg = "";
  let isNewOrgDialogOpen = false;

  function selectHandler() {
    open = false;
    onSelect(selectedOrg);
  }

  function handleCreateOrg(orgName: string) {
    selectedOrg = orgName;
    isNewOrgDialogOpen = false;
    eventBus.emit("notification", {
      message: `Created organization ${orgName}`,
    });
  }
</script>

<div class="text-xl">Select an organization</div>
<div class="text-base text-gray-500">
  Choose an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org"
    target="_blank">See docs</a
  >
</div>
<!-- w-[400px] Needed for tailwind to compile for this -->
<Select
  bind:value={selectedOrg}
  id="deploy-target-org"
  label=""
  placeholder="Select organization"
  options={orgs.map((o) => ({ value: o, label: o }))}
  width={400}
  sameWidth
>
  <div slot="extra-dropdown-content">
    <SelectSeparator />
    <SelectItem
      value="__rill_new_org"
      on:click={() => (isNewOrgDialogOpen = true)}
      class="text-[12px] gap-x-2 items-start"
    >
      + Create organization
    </SelectItem>
  </div>
</Select>
<Button wide type="primary" on:click={selectHandler}>Continue</Button>

<Dialog.Root bind:open={isNewOrgDialogOpen}>
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Title>Create a new organization</Dialog.Title>

    <CreateNewOrgForm onCreate={handleCreateOrg} />

    <Dialog.Footer class="gap-x-2">
      <Button large type="text" on:click={() => (isNewOrgDialogOpen = false)}>
        Cancel
      </Button>
      <Button large type="primary" submitForm form={CreateNewOrgFormId}>
        Continue
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
