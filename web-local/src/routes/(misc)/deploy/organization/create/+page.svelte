<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { getDeployOrGithubRouteGetter } from "@rilldata/web-common/features/project/deploy/route-utils.ts";

  const deployRouteGetter = getDeployOrGithubRouteGetter();
  $: ({ isLoading, getter: deployRouteGetterFunc } = $deployRouteGetter);

  function selectOrg(orgName: string) {
    // This navigation gets cancelled if we do not have `setTimeout` here.
    // TODO: investigate why
    setTimeout(() => void goto(deployRouteGetterFunc(orgName)));
  }
</script>

<div class="text-xl">Let’s create your first organization</div>
<div class="text-base text-gray-500">
  Create an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org/create"
    target="_blank">See docs</a
  >
</div>

<div class="text-left">
  <CreateNewOrgForm onCreate={selectOrg} size="xl" />
</div>
<Button
  wide
  forcedStyle="min-width:500px !important;"
  type="primary"
  submitForm
  form={CreateNewOrgFormId}
  loading={isLoading}
  disabled={isLoading}
>
  Continue
</Button>
