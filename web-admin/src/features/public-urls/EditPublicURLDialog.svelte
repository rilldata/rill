<script lang="ts">
  import { page } from "$app/state";
  import {
    createAdminServiceUpdateMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
    type AdminServiceUpdateMagicAuthTokenBody,
    type V1MagicAuthToken,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { LockIcon } from "lucide-svelte";
  import PasswordInput from "./PasswordInput.svelte";
  import SavedPassword from "./SavedPassword.svelte";
  import { isValidPublicURLPassword } from "./password-policy";

  let {
    open = $bindable(false),
    token,
  }: { open?: boolean; token: V1MagicAuthToken } = $props();

  const formId = "edit-public-url-form";
  const queryClient = useQueryClient();
  const updateMagicAuthToken = createAdminServiceUpdateMagicAuthToken();

  let displayName = $state("");
  let passwordAction = $state<"keep" | "set" | "remove">("keep");
  let newPassword = $state("");
  let apiError = $state("");
  // Shown once after saving: the password is stored as a hash and can't be retrieved later.
  let savedPassword = $state<string | null>(null);

  // Start from the token's current values each time the dialog opens.
  $effect(() => {
    if (!open) return;
    displayName = token.displayName ?? "";
    passwordAction = "keep";
    newPassword = "";
    apiError = "";
    savedPassword = null;
  });

  function setNewPassword(password: string) {
    newPassword = password;
    if (token.passwordProtected) passwordAction = password ? "set" : "keep";
  }

  function setPasswordAction(action: "keep" | "set" | "remove") {
    passwordAction = action;
    newPassword = "";
  }

  let hasChanges = $derived(
    displayName !== (token.displayName ?? "") || passwordAction !== "keep",
  );
  let canSave = $derived(
    !$updateMagicAuthToken.isPending &&
      hasChanges &&
      displayName.trim() !== "" &&
      (passwordAction !== "set" || isValidPublicURLPassword(newPassword)),
  );

  async function save() {
    if (!canSave) return;
    const data: AdminServiceUpdateMagicAuthTokenBody = {};
    if (displayName !== (token.displayName ?? "")) {
      data.displayName = displayName;
    }
    if (passwordAction === "set") data.password = newPassword;
    if (passwordAction === "remove") data.removePassword = true;

    try {
      await $updateMagicAuthToken.mutateAsync({
        tokenId: token.id ?? "",
        data,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListMagicAuthTokensQueryKey(
          page.params.organization ?? "",
          page.params.project ?? "",
        ),
      });
      eventBus.emit("notification", {
        message: m.public_url_updated_notification(),
      });
      if (data.password) {
        savedPassword = data.password;
      } else {
        open = false;
      }
    } catch (error) {
      const typedError = error as HTTPError;
      apiError = typedError.response?.data?.message ?? typedError.message;
    }
  }
</script>

<Dialog bind:open>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>{m.public_url_edit_title()}</DialogTitle>
    </DialogHeader>
    {#if savedPassword}
      <div class="flex flex-col gap-y-2">
        <p class="text-sm font-semibold text-fg-primary">
          {m.public_url_new_password_heading()}
        </p>
        <SavedPassword password={savedPassword} />
      </div>
      <DialogFooter>
        <Button
          type="primary"
          onClick={() => {
            open = false;
          }}
        >
          {m.public_url_done_button()}
        </Button>
      </DialogFooter>
    {:else}
      <form
        id={formId}
        class="flex flex-col gap-y-4"
        onsubmit={(e) => {
          e.preventDefault();
          void save();
        }}
      >
        <Input
          id="edit-public-url-label"
          bind:value={displayName}
          label={m.public_url_edit_label()}
          placeholder={m.public_url_label_placeholder()}
        />

        {#if token.passwordProtected}
          <div class="flex flex-col gap-y-2">
            <p class="flex items-center gap-x-1 text-xs text-fg-primary">
              <LockIcon size={12} class="shrink-0" />
              {m.public_url_password_is_set()}
            </p>
            <PasswordInput
              id="edit-public-url-password"
              bind:value={() => newPassword, setNewPassword}
              placeholder={m.public_url_set_new_password()}
              disabled={passwordAction === "remove"}
            />
            <div class="flex items-center gap-x-2">
              <Switch
                small
                id="remove-password"
                checked={passwordAction === "remove"}
                onCheckedChange={(checked) =>
                  setPasswordAction(checked ? "remove" : "keep")}
              />
              <Label class="text-xs" for="remove-password"
                >{m.public_url_remove_password()}</Label
              >
            </div>
            <p class="text-xs text-fg-secondary">
              {m.public_url_password_change_notice()}
            </p>
          </div>
        {:else}
          <div class="flex flex-col gap-y-2">
            <div class="flex items-center gap-x-2">
              <Switch
                small
                id="require-password"
                checked={passwordAction === "set"}
                onCheckedChange={(checked) =>
                  setPasswordAction(checked ? "set" : "keep")}
              />
              <Label class="text-xs" for="require-password"
                >{m.public_url_require_password()}</Label
              >
            </div>
            {#if passwordAction === "set"}
              <PasswordInput
                id="edit-public-url-password"
                bind:value={() => newPassword, setNewPassword}
              />
            {/if}
          </div>
        {/if}

        {#if apiError}
          <div class="text-xs text-red-500">{apiError}</div>
        {/if}
      </form>
      <DialogFooter>
        <Button
          type="tertiary"
          onClick={() => {
            open = false;
          }}
        >
          {m.public_url_cancel_button()}
        </Button>
        <Button type="primary" form={formId} submitForm disabled={!canSave}>
          {m.public_url_save_button()}
        </Button>
      </DialogFooter>
    {/if}
  </DialogContent>
</Dialog>
