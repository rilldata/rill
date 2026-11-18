<script lang="ts">
  import {
    createAdminServiceListProjectWhitelistedDomains,
    createAdminServiceRemoveProjectWhitelistedDomain,
    getAdminServiceListProjectWhitelistedDomainsQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { capitalize } from "@rilldata/web-common/components/table/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";

  // Lists the project's domain access rules, which add users by email domain without an invite,
  // so the Share dialog does not look invite-only while one is active.
  // Only project admins may list and remove these rules.

  export let organization: string;
  export let project: string;
  export let enabled: boolean;

  const queryClient = useQueryClient();
  const removeDomain = createAdminServiceRemoveProjectWhitelistedDomain();

  $: domainsQuery = createAdminServiceListProjectWhitelistedDomains(
    organization,
    project,
    { query: { enabled, refetchOnMount: true } },
  );
  $: domains = $domainsQuery.data?.domains ?? [];

  async function remove(domain: string) {
    try {
      await $removeDomain.mutateAsync({ org: organization, project, domain });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectWhitelistedDomainsQueryKey(
          organization,
          project,
        ),
      });
      eventBus.emit("notification", {
        message: m.users_domain_access_removed(),
      });
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message:
          (e as AxiosError<RpcStatus>).response?.data?.message ??
          (e as Error).message,
      });
    }
  }
</script>

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
        {m.users_domain_access_joins_as({ role: capitalize(role ?? "") })}
      </span>
    </div>
    <div class="grow"></div>
    <Button
      type="text"
      small
      loading={$removeDomain.isPending &&
        $removeDomain.variables?.domain === domain}
      onClick={() => remove(domain ?? "")}
    >
      {m.users_remove()}
    </Button>
  </div>
{/each}
