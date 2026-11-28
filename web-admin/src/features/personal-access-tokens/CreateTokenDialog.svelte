<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { tryCopyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { DateTime } from "luxon";
  import { onDestroy } from "svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import ExpirationSelect from "./ExpirationSelect.svelte";
  import {
    DEFAULT_TOKEN_EXPIRATION,
    formatTokenDate,
    issuePersonalAccessToken,
    TOKEN_EXPIRATIONS,
    TOKEN_NAME_MAX_LENGTH,
  } from "./personal-access-tokens";

  let { open = $bindable(false) }: { open?: boolean } = $props();

  const queryClient = useQueryClient();

  let apiError = $state("");
  // The secret is only in the issue response, so it is shown once and dropped when the dialog closes.
  let issued = $state<{
    token: string;
    name: string;
    expiresOn: DateTime | null;
  } | null>(null);
  let tokenElement = $state<HTMLElement | null>(null);
  let copyState = $state<"idle" | "copied" | "failed">("idle");
  let copyTimeout: ReturnType<typeof setTimeout> | undefined;

  const schema = yup(
    object({
      name: string()
        .trim()
        .required(m.access_tokens_name_required())
        .max(
          TOKEN_NAME_MAX_LENGTH,
          m.access_tokens_name_too_long({ max: TOKEN_NAME_MAX_LENGTH }),
        ),
      expiration: string().oneOf(TOKEN_EXPIRATIONS).required(),
      // The day picked for the "custom" expiration, as yyyy-MM-dd.
      customDate: string(),
    }),
  );

  const { form, formId, errors, enhance, submit, submitting, reset } =
    superForm(
      defaults(
        { name: "", expiration: DEFAULT_TOKEN_EXPIRATION, customDate: "" },
        schema,
      ),
      {
        SPA: true,
        validators: schema,
        // The form is reset when the dialog closes; keep the values after an API error.
        resetForm: false,
        async onUpdate({ form }) {
          if (!form.valid) return;
          apiError = "";
          const name = form.data.name.trim();
          try {
            const { token, expiresOn } = await issuePersonalAccessToken(
              queryClient,
              name,
              form.data.expiration,
              form.data.customDate ?? "",
            );
            issued = { token, name, expiresOn };
          } catch (e) {
            apiError =
              (e as HTTPError).response?.data?.message ??
              m.access_tokens_create_failed();
          }
        },
      },
    );

  async function copyToken() {
    if (!issued) return;
    clearTimeout(copyTimeout);
    if (await tryCopyToClipboard(issued.token)) {
      copyState = "copied";
      copyTimeout = setTimeout(() => (copyState = "idle"), 2_000);
      return;
    }
    // Clipboard access can be denied: select the token so it can be copied by hand.
    copyState = "failed";
    if (tokenElement) {
      const range = document.createRange();
      range.selectNodeContents(tokenElement);
      window.getSelection()?.removeAllRanges();
      window.getSelection()?.addRange(range);
    }
  }

  function close() {
    open = false;
  }

  onDestroy(() => clearTimeout(copyTimeout));
</script>

<!-- Outside clicks never close it: one could lose a token that hasn't been copied yet. -->
<Dialog.Root
  bind:open
  onOpenChangeComplete={(isOpen) => {
    if (isOpen) return;
    reset();
    apiError = "";
    issued = null;
    clearTimeout(copyTimeout);
    copyState = "idle";
  }}
>
  <!-- While the request runs it can't be closed: the server may still issue a token whose secret would be lost. -->
  <Dialog.Content
    noClose={$submitting}
    escapeKeydownBehavior={$submitting ? "ignore" : "close"}
    interactOutsideBehavior="ignore"
  >
    {#if !issued}
      <Dialog.Header>
        <Dialog.Title>{m.access_tokens_create_title()}</Dialog.Title>
        <Dialog.Description>
          {m.access_tokens_create_description()}
        </Dialog.Description>
      </Dialog.Header>

      <form id={$formId} class="flex flex-col gap-y-4" use:enhance>
        <Input
          id="access-token-name"
          label={m.common_name()}
          placeholder={m.access_tokens_name_placeholder()}
          bind:value={$form.name}
          errors={$errors.name}
          alwaysShowError
          claimFocusOnMount
          onEnter={() => submit()}
        />
        <ExpirationSelect
          id="access-token-expiration"
          bind:expiration={$form.expiration}
          bind:customDate={$form.customDate}
        />
        {#if apiError}
          <p class="text-destructive text-sm">{apiError}</p>
        {/if}
      </form>

      <Dialog.Footer>
        <Button type="secondary" onClick={close} disabled={$submitting}>
          {m.common_cancel()}
        </Button>
        <Button
          type="primary"
          submitForm
          form={$formId}
          loading={$submitting}
          loadingCopy={m.access_tokens_creating()}
          disabled={$submitting}
        >
          {m.access_tokens_create_submit()}
        </Button>
      </Dialog.Footer>
    {:else}
      <Dialog.Header>
        <Dialog.Title>{m.access_tokens_copy_title()}</Dialog.Title>
        <Dialog.Description>
          {m.access_tokens_copy_description()}
        </Dialog.Description>
      </Dialog.Header>

      <div class="flex flex-col gap-y-2">
        <div class="flex items-center gap-x-2">
          <!-- One line cut with an ellipsis; a click still selects the whole token for a manual copy. -->
          <code
            bind:this={tokenElement}
            class="flex-1 min-w-0 truncate select-all rounded-[2px] border bg-surface-background px-2 py-1.5 font-mono text-xs text-fg-primary"
          >
            {issued.token}
          </code>
          <Button type="secondary" onClick={copyToken}>
            {copyState === "copied"
              ? m.access_tokens_copied()
              : m.access_tokens_copy()}
          </Button>
        </div>
        {#if copyState === "failed"}
          <p class="text-xs text-fg-secondary" role="status">
            {m.access_tokens_copy_failed()}
          </p>
        {/if}
        <p class="text-xs text-fg-secondary">
          {issued.expiresOn
            ? m.access_tokens_created_expires({
                name: issued.name,
                date: formatTokenDate(issued.expiresOn),
              })
            : m.access_tokens_created_never_expires({ name: issued.name })}
          {m.access_tokens_usage_hint()}
        </p>
      </div>

      <Dialog.Footer>
        <Button type="primary" onClick={close}>{m.access_tokens_done()}</Button>
      </Dialog.Footer>
    {/if}
  </Dialog.Content>
</Dialog.Root>
