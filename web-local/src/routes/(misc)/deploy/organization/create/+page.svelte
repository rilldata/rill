<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import CreateNewOrgForm from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { CreateNewOrgFormId } from "@rilldata/web-common/features/organization/CreateNewOrgForm.svelte";
  import { getDeployOrGithubRouteGetter } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import {
    createLocalServiceCreateOrganization,
    getLocalServiceGetCurrentUserQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  const deployRouteGetter = getDeployOrGithubRouteGetter();
  $: ({ isLoading, getter: deployRouteGetterFunc } = $deployRouteGetter);

  const orgCreator = createLocalServiceCreateOrganization();

  async function createOrg(name: string, displayName: string) {
    await $orgCreator.mutateAsync({
      name,
      displayName,
    });

    await queryClient.invalidateQueries({
      queryKey: getLocalServiceGetCurrentUserQueryKey(),
    });

    // This navigation gets cancelled because of form submission.
    setTimeout(() => void goto(deployRouteGetterFunc(name)));
  }
</script>

<div class="text-xl">Let’s create your first organization</div>
<div class="text-base text-fg-secondary">
  Create an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org/create"
    target="_blank">See docs</a
  >
</div>

<div class="text-left">
  <CreateNewOrgForm {createOrg} size="xl" />
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
