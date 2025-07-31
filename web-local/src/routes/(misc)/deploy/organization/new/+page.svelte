<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";

  function selectOrg(orgName: string) {
    setTimeout(() => void goto(`/deploy/new?org=${orgName}`));
  }
</script>

<div class="text-xl">Let’s create your first organization</div>
<div class="text-base text-gray-500">
  Create an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org/create"
    target="_blank">See docs</a
  >
</div>

<CreateNewOrgForm onCreate={selectOrg} size="xl" />
<Button
  wide
  forcedStyle="min-width:500px !important;"
  type="primary"
  submitForm
  form={CreateNewOrgFormId}
>
  Continue
</Button>
