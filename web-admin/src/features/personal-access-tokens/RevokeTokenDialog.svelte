<script lang="ts">
  import {
    createAdminServiceRevokeUserAuthToken,
    getAdminServiceListUserAuthTokensQueryKey,
    type V1UserAuthToken,
  } from "@rilldata/web-admin/client";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";

  let {
    open = $bindable(false),
    token,
  }: { open?: boolean; token: V1UserAuthToken } = $props();

  const queryClient = useQueryClient();
  const revokeToken = createAdminServiceRevokeUserAuthToken();

  let apiError = $state("");

  async function revoke() {
    apiError = "";
    try {
      await $revokeToken.mutateAsync({ tokenId: token.id ?? "" });
      open = false;
      eventBus.emit("notification", { message: m.access_tokens_revoked() });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListUserAuthTokensQueryKey("current"),
      });
    } catch (e) {
      const error = e as HTTPError;
      apiError = m.access_tokens_revoke_failed({
        message: error.response?.data?.message ?? error.message,
      });
    }
  }
</script>

<!-- Closing is blocked while the request runs, so its result can't land on another token's dialog. -->
<AlertDialog bind:open>
  <AlertDialogContent
    noCancel={$revokeToken.isPending}
    escapeKeydownBehavior={$revokeToken.isPending ? "ignore" : "close"}
  >
    <AlertDialogHeader>
      <AlertDialogTitle>
        {m.access_tokens_revoke_title({ name: token.displayName ?? "" })}
      </AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1 flex flex-col gap-y-2">
          <span>{m.access_tokens_revoke_description()}</span>
          <span class="font-mono text-xs">{token.prefix}…</span>
          {#if apiError}
            <span class="text-destructive">{apiError}</span>
          {/if}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => (open = false)}
        disabled={$revokeToken.isPending}
      >
        {m.common_cancel()}
      </Button>
      <Button
        type="destructive"
        onClick={revoke}
        loading={$revokeToken.isPending}
        disabled={$revokeToken.isPending}
      >
        {m.access_tokens_revoke_confirm()}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
