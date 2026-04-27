<script lang="ts">
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { createAdminServiceListDeployments } from "@rilldata/web-admin/client";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import { injectBranchIntoPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let deploymentsQuery = $derived(
    createAdminServiceListDeployments(organization, project, {}),
  );
  let prodDeployment = $derived(
    $deploymentsQuery.data?.deployments?.find(
      (d) => !d.editable && d.environment === "prod",
    ),
  );
  let editableDeployment = $derived(
    $deploymentsQuery.data?.deployments?.find((d) => d.editable),
  );
</script>

{#if prodDeployment || !editableDeployment}
  <!-- Primary deployment exists and is stopped/stopping. -->
  <RedeployProjectCta {organization} {project} />
{:else if editableDeployment}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <CtaHeader variant="bold">
        This project hasn’t been published yet. What would you like to do next?
      </CtaHeader>
      <div class="flex flex-row gap-2 justify-start">
        <Button
          type="secondary"
          href={injectBranchIntoPath(
            `/${organization}/${project}/-/edit`,
            editableDeployment.branch,
          )}
        >
          Continue editing
        </Button>
        <Button type="primary">Publish (TODO)</Button>
      </div>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
