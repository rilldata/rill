<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { getDeployOrGithubRouteGetter } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { createLocalServiceGetCurrentUser } from "@rilldata/web-common/runtime-client/local-service.ts";

  const user = createLocalServiceGetCurrentUser();

  let selectedOrg = "";
  let isNewOrgDialogOpen = false;

  const deployRouteGetter = getDeployOrGithubRouteGetter();
  $: ({ isLoading, getter: deployRouteGetterFunc } = $deployRouteGetter);
  $: createProjectUrl = deployRouteGetterFunc(selectedOrg);

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
  onAddNew={() => (isNewOrgDialogOpen = true)}
  addNewLabel="+ Create organization"
  width={400}
  sameWidth
/>

<Button
  wide
  type="primary"
  href={createProjectUrl}
  loading={isLoading}
  disabled={!selectedOrg || isLoading}
>
  Deploy as a new project
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
