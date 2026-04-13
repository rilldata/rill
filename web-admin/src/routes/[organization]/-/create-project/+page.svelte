<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { onMount } from "svelte";
  import {
    createAdminServiceCreateProject,
    createAdminServiceListProjectsForOrganization,
  } from "@rilldata/web-admin/client";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";

  let organization = $derived(page.params.organization);

  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization, undefined),
  );

  const createProjectMutation = createAdminServiceCreateProject();

  async function createProject() {
    await waitUntil(() => !$projectsQuery.isPending);
    const projectNames =
      $projectsQuery.data?.projects?.map((p) => p.name) ?? [];
    const newProjectName = getName("project", projectNames);

    const createProjectResp = await $createProjectMutation.mutateAsync({
      org: organization,
      data: {
        project: newProjectName,
        generateManagedGit: true,
        prodSlots: "4",
      },
    });
    const frontendUrl = createProjectResp.project?.frontendUrl;
    if (!frontendUrl) return;
    await goto(`${frontendUrl}/-/welcome`);
  }

  onMount(() => void createProject());
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    <CtaHeader variant="bold">
      Hang tight! We're creating your project...
    </CtaHeader>
  </CtaContentContainer>
</CtaLayoutContainer>
