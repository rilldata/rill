<!--
  Shown in place of a public URL's dashboard until the recipient enters the URL's password.
  A correct password returns a session token that the root layout then uses as the bearer token.
-->
<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { LockIcon } from "lucide-svelte";
  import { setPublicURLSession } from "./password-session";
  import { adminServiceUnlockMagicAuthToken } from "./unlock-magic-auth-token";

  let { token }: { token: string } = $props();

  // `Button` only submits the form it's given by id (without one it renders an empty `form` attribute,
  // which detaches it from the form). `Input` blurs itself on Enter before the browser would submit
  // the form, so Enter submits through `onEnter` instead.
  const formId = "public-url-password-form";

  let password = $state("");
  let submitting = $state(false);
  let error = $state("");

  async function unlock() {
    if (!password || submitting) return;
    submitting = true;
    error = "";
    try {
      const { sessionToken, expiresOn } =
        await adminServiceUnlockMagicAuthToken(token, password);
      if (!sessionToken || !expiresOn) {
        throw new Error("Unlocking the public URL did not return a session");
      }
      setPublicURLSession(token, { token: sessionToken, expiresOn });
      // Re-run the root layout's loader, which now finds the session and loads the project.
      await invalidateAll();
    } catch (e) {
      const status = (e as HTTPError).response?.status;
      if (status === 403) {
        error = m.public_url_gate_incorrect();
      } else if (status === 429) {
        error = m.public_url_gate_too_many_attempts();
      } else {
        error =
          (e as HTTPError).response?.data?.message ?? (e as Error).message;
      }
    } finally {
      submitting = false;
    }
  }
</script>

<div class="flex flex-1 items-center justify-center p-8">
  <form
    id={formId}
    class="flex w-full max-w-sm flex-col items-center gap-y-4 rounded-lg border bg-surface-background p-8 text-center"
    onsubmit={(e) => {
      e.preventDefault();
      void unlock();
    }}
  >
    <div
      class="grid size-10 place-items-center rounded-full bg-surface-secondary"
    >
      <LockIcon size={18} class="text-fg-secondary" />
    </div>
    <h1 class="text-base font-semibold text-fg-primary">
      {m.public_url_gate_title()}
    </h1>
    <p class="text-sm text-fg-secondary">{m.public_url_gate_description()}</p>
    <Input
      id="public-url-password"
      bind:value={password}
      secret
      placeholder={m.public_url_password_placeholder()}
      claimFocusOnMount
      errors={error}
      alwaysShowError
      onEnter={() => void unlock()}
    />
    <Button
      type="primary"
      form={formId}
      submitForm
      wide
      disabled={!password || submitting}
      loading={submitting}
    >
      {m.public_url_gate_submit()}
    </Button>
  </form>
</div>
