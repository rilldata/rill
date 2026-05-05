<script lang="ts">
  import { createAdminServiceListDeployments } from "@rilldata/web-admin/client";
  import { requestSkipBranchInjection } from "@rilldata/web-admin/features/branches/branch-utils";
  import { isProdDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { LogOut } from "lucide-svelte";

  export let organization: string;
  export let project: string;

  const deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
  );

  // While the query is loading, default `hasProdDeployment` to false so a
  // click during the loading window routes to the org page (the safe
  // default for an unpublished project). TanStack typically has this query
  // warm from PublishPopover, so the loading window is rare.
  $: hasProdDeployment =
    $deploymentsQuery.data?.deployments?.some(isProdDeployment) ?? false;
  $: href = hasProdDeployment
    ? `/${organization}/${project}`
    : `/${organization}`;

  // The project layout's beforeNavigate hook re-injects the active @branch
  // into project-scoped URLs. Exit wants the bare URL, so request a skip
  // before SvelteKit intercepts the anchor click. Harmless when href is
  // the org page (the layout doesn't inject for non-project URLs).
  function handleClick() {
    requestSkipBranchInjection();
  }
</script>

<Tooltip distance={8}>
  <Button type="secondary" {href} onClick={handleClick}>
    <LogOut size="14" />
    Exit
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">
      {hasProdDeployment ? "Return to project home" : "Return to organization"}
    </span>
  </TooltipContent>
</Tooltip>
