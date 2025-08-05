<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { SelectSeparator } from "@rilldata/web-common/components/select";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import {
    getCreateProjectRoute,
    getOverwriteProjectRoute,
  } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { createLocalServiceGetCurrentUser } from "@rilldata/web-common/runtime-client/local-service.ts";

  const user = createLocalServiceGetCurrentUser();

  let selectedOrg = "";
  let isNewOrgDialogOpen = false;

  $: createProjectUrl = getCreateProjectRoute(selectedOrg);
  $: overwriteProjectUrl = getOverwriteProjectRoute(selectedOrg);

  $: orgOptions =
    $user.data?.rillUserOrgs?.map((o) => ({ value: o, label: o })) ?? [];

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
  ariaLabel="Select organization"
  placeholder="Select organization"
  options={orgOptions}
  width={400}
  sameWidth
>
  <div slot="additional-dropdown-content" let:close>
    <SelectSeparator />
    <button
      on:click={() => {
        isNewOrgDialogOpen = true;
        close();
      }}
      class="w-full cursor-pointer select-none rounded-sm py-1.5 px-2 text-left hover:bg-accent"
    >
      + Create organization
    </button>
  </div>
</Select>

<Button wide type="primary" href={createProjectUrl} disabled={!selectedOrg}>
  Deploy as a new project
</Button>
<Button
  wide
  type="ghost"
  href={overwriteProjectUrl}
  disabled={!selectedOrg}
  class="-mt-2"
>
  Or overwrite an existing project
</Button>

<Dialog.Root bind:open={isNewOrgDialogOpen}>
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content noClose>
    <Dialog.Title>Create a new organization</Dialog.Title>

    <CreateNewOrgForm onCreate={handleCreateOrg} size="lg" />

    <Dialog.Footer class="gap-x-2">
      <Button large type="text" onClick={() => (isNewOrgDialogOpen = false)}>
        Cancel
      </Button>
      <Button
        large
        type="primary"
        submitForm
        form={CreateNewOrgFormId}
        label="Create new org"
      >
        Continue
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
