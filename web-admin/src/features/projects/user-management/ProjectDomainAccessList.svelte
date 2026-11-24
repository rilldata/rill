<script lang="ts">
  import {
    createAdminServiceListProjectWhitelistedDomains,
    createAdminServiceRemoveProjectWhitelistedDomain,
    getAdminServiceListProjectWhitelistedDomainsQueryKey,
  } from "@rilldata/web-admin/client";
  import { getProjectRolesOptions } from "@rilldata/web-admin/features/projects/user-management/constants";
  import { AlertDialogConfirmation } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";

  // Lists the project's domain access rules, which add users by email domain without an invite,
  // so the Share dialog does not look invite-only while one is active.
  // Only project admins may list and remove these rules.

  export let organization: string;
  export let project: string;
  export let enabled: boolean;

  const queryClient = useQueryClient();
  const removeDomain = createAdminServiceRemoveProjectWhitelistedDomain();
  // Set before the request starts, so a second click can't send another one,
  // and cleared once the list has refetched without the removed domain.
  let removing: string | undefined;
  // Removal asks first: an admin can't add the rule back from the Share dialog.
  // The rule stays set while the dialog closes so its text doesn't change mid-animation.
  let confirmOpen = false;
  let confirmRule = { domain: "", role: "" };
  const roleLabels = new Map<string, string>(
    getProjectRolesOptions().map(({ value, label }) => [value, label]),
  );

  $: domainsQuery = createAdminServiceListProjectWhitelistedDomains(
    organization,
    project,
    { query: { enabled, refetchOnMount: true } },
  );
  $: domains = $domainsQuery.data?.domains ?? [];

  async function remove(domain: string) {
    if (removing) return;
    removing = domain;
    try {
      await $removeDomain.mutateAsync({ org: organization, project, domain });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectWhitelistedDomainsQueryKey(
          organization,
          project,
        ),
      });
      eventBus.emit("notification", {
        message: m.users_domain_access_removed({ domain }),
      });
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(e),
      });
    } finally {
      removing = undefined;
    }
  }
</script>

{#if $domainsQuery.isError}
  <!-- An empty list would look invite-only, so say that the rules could not be loaded. -->
  <div class="flex flex-row items-center gap-x-2 py-2 pl-2">
    <span class="text-xs text-red-600">
      {m.users_domain_access_load_failed()}
    </span>
    <div class="grow"></div>
    <Button type="text" small onClick={() => $domainsQuery.refetch()}>
      {m.users_retry()}
    </Button>
  </div>
{/if}
{#each domains as { domain, role } (domain)}
  <div
    class="flex flex-row items-center gap-x-2 py-2 pl-2"
    data-testid="project-domain-access"
  >
    <div
      class="h-7 w-7 shrink-0 rounded-sm flex items-center justify-center bg-surface-hover text-sm font-semibold text-fg-secondary"
    >
      @
    </div>
    <div class="flex flex-col text-left min-w-0">
      <span class="text-sm font-medium text-fg-primary truncate">
        {m.users_domain_access_anyone({ domain: domain ?? "" })}
      </span>
      <span class="text-xs text-fg-secondary">
        {m.users_domain_access_joins_as({
          role: roleLabels.get(role ?? "") ?? capitalize(role ?? ""),
        })}
      </span>
    </div>
    <div class="grow"></div>
    <!-- Same box and text as the role dropdowns above, so the column lines up.
         The row being removed keeps full opacity and a spinner just left of the box. -->
    <button
      type="button"
      class="relative rounded-sm mr-[10px] min-w-[72px] px-2 py-1 text-left hover:bg-surface-hover disabled:cursor-not-allowed"
      class:opacity-50={!!removing && removing !== domain}
      disabled={!!removing}
      aria-busy={removing === domain}
      onclick={() => {
        confirmRule = { domain: domain ?? "", role: role ?? "" };
        confirmOpen = true;
      }}
    >
      {#if removing === domain}
        <span class="absolute -left-4 top-1/2 -translate-y-1/2">
          <LoadingSpinner size="12px" />
        </span>
      {/if}
      {m.users_remove()}
    </button>
  </div>
{/each}

<AlertDialogConfirmation
  open={confirmOpen}
  onOpenChange={(open: boolean) => {
    confirmOpen = open;
  }}
  title={m.users_domain_access_remove_confirm_title({
    domain: confirmRule.domain,
  })}
  description={m.users_domain_access_remove_confirm_desc({
    domain: confirmRule.domain,
    organization,
    project,
    role: confirmRule.role,
  })}
  confirmLabel={m.users_yes_remove()}
  confirmType="destructive"
  onConfirm={() => {
    confirmOpen = false;
    void remove(confirmRule.domain);
  }}
/>
