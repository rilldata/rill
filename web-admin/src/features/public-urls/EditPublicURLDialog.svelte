<script lang="ts">
  import { page } from "$app/state";
  import {
    createAdminServiceUpdateMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
    type AdminServiceUpdateMagicAuthTokenBody,
    type V1MagicAuthToken,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
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
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { LockIcon } from "lucide-svelte";
  import GeneratePasswordButton from "./GeneratePasswordButton.svelte";

  let {
    open = $bindable(false),
    token,
  }: { open?: boolean; token: V1MagicAuthToken } = $props();

  const formId = "edit-public-url-form";
  const queryClient = useQueryClient();
  const updateMagicAuthToken = createAdminServiceUpdateMagicAuthToken();

  let displayName = $state("");
  // For a URL without a password, toggles adding one.
  let requirePassword = $state(false);
  // Empty keeps the current password.
  let newPassword = $state("");
  let removePassword = $state(false);
  let apiError = $state("");
  // Shown once after saving: the password is stored as a hash and can't be retrieved later.
  let savedPassword = $state<string | null>(null);
  let passwordCopied = $state(false);

  // Start from the token's current values each time the dialog opens.
  $effect(() => {
    if (!open) return;
    displayName = token.displayName ?? "";
    requirePassword = !!token.passwordProtected;
    newPassword = "";
    removePassword = false;
    apiError = "";
    savedPassword = null;
  });

  // The new password starts empty: the user either types one or generates one with the button below the field.
  $effect(() => {
    if (removePassword || !requirePassword) newPassword = "";
  });

  let passwordInvalid = $derived(
    newPassword !== "" && (newPassword.length < 8 || newPassword.length > 72),
  );
  let passwordChanged = $derived(
    token.passwordProtected
      ? removePassword || newPassword !== ""
      : requirePassword && newPassword !== "",
  );
  let hasChanges = $derived(
    displayName !== (token.displayName ?? "") || passwordChanged,
  );
  let canSave = $derived(
    !$updateMagicAuthToken.isPending &&
      hasChanges &&
      displayName.trim() !== "" &&
      !passwordInvalid &&
      (token.passwordProtected || !requirePassword || newPassword !== ""),
  );

  async function save() {
    const data: AdminServiceUpdateMagicAuthTokenBody = {};
    if (displayName !== (token.displayName ?? "")) {
      data.displayName = displayName;
    }
    if (token.passwordProtected) {
      if (removePassword) data.removePassword = true;
      else if (newPassword) data.password = newPassword;
    } else if (requirePassword && newPassword) {
      data.password = newPassword;
    }

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

  function onCopyPassword() {
    if (!savedPassword) return;
    copyToClipboard(savedPassword, "Password copied to clipboard", false);
    passwordCopied = true;

    setTimeout(() => {
      passwordCopied = false;
    }, 2_000);
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
        <p class="text-xs text-fg-secondary">
          {m.public_url_password_save_notice()}
        </p>
        <div
          class="flex items-center justify-between gap-x-2 rounded border bg-surface-secondary px-2 py-1"
        >
          <code
            class="truncate font-mono text-xs"
            data-public-url-password={savedPassword}>{savedPassword}</code
          >
          <Button type="text" compact onClick={onCopyPassword}>
            {#if passwordCopied}
              <Check size="14px" />
              {m.public_url_copied_password_button()}
            {:else}
              {m.public_url_copy_password_button()}
            {/if}
          </Button>
        </div>
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
            <div class="flex flex-col items-start gap-y-1">
              <Input
                id="edit-public-url-password"
                bind:value={newPassword}
                secret
                placeholder={m.public_url_set_new_password()}
                disabled={removePassword}
                errors={passwordInvalid ? m.public_url_password_length() : null}
                alwaysShowError
              />
              <GeneratePasswordButton
                disabled={removePassword}
                onGenerate={(password) => {
                  newPassword = password;
                }}
              />
            </div>
            <div class="flex items-center gap-x-2">
              <Switch
                small
                id="remove-password"
                bind:checked={removePassword}
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
                bind:checked={requirePassword}
              />
              <Label class="text-xs" for="require-password"
                >{m.public_url_require_password()}</Label
              >
            </div>
            {#if requirePassword}
              <div class="flex flex-col items-start gap-y-1">
                <Input
                  id="edit-public-url-password"
                  bind:value={newPassword}
                  secret
                  placeholder={m.public_url_password_placeholder()}
                  errors={passwordInvalid
                    ? m.public_url_password_length()
                    : null}
                  alwaysShowError
                />
                <GeneratePasswordButton
                  onGenerate={(password) => {
                    newPassword = password;
                  }}
                />
              </div>
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
