<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { createAdminServiceCreateOrganization } from "@rilldata/web-admin/client";

  const createOrgMutation = createAdminServiceCreateOrganization();
  $: ({ isFetching } = $createOrgMutation);

  async function createOrg(name: string, displayName: string) {
    await $createOrgMutation.mutateAsync({
      data: {
        name,
        displayName,
      },
    });

    // This navigation gets cancelled if we do not have `setTimeout` here.
    setTimeout(() => void goto(`/${name}`));
  }
</script>

<div class="flex flex-col gap-4 mx-auto w-fit">
  <RillLogoSquareNegative size="36px" />
  <div class="text-2xl font-extrabold text-fg-accent text-center">
    Create an organization
  </div>

  <div
    class="flex flex-col gap-6 text-left p-6 border rounded-md bg-surface-overlay"
  >
    <div>
      <div class="text-base font-semibold">Name your organization</div>
      <div class="text-sm text-fg-muted">
        You can change the name in organization setting.
      </div>
    </div>
    <CreateNewOrgForm {createOrg} size="xl" />
    <div class="w-full flex justify-end">
      <Button
        type="primary"
        submitForm
        form={CreateNewOrgFormId}
        loading={isFetching}
      >
        Continue
      </Button>
    </div>
  </div>
</div>
