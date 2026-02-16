<script lang="ts">
  import { createForm } from "svelte-forms-lib";
  import { Dialog } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { createUserTokenMutation } from "./token-queries";
  import { createEventDispatcher } from "svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Copy from "@rilldata/web-common/components/icons/Copy.svelte";

  export let open = false;

  const dispatch = createEventDispatcher();

  const createToken = createUserTokenMutation();

  let step: "form" | "display" = "form";
  let createdToken = "";
  let copied = false;
  let apiError = "";

  const EXPIRATION_OPTIONS = [
    { value: "30", label: "30 days" },
    { value: "60", label: "60 days" },
    { value: "90", label: "90 days" },
    { value: "custom", label: "Custom date" },
    { value: "none", label: "No expiration" },
  ];

  let name = "";
  let description = "";
  let expirationOption = "90";
  let customDate = "";
  let nameError = "";
  let customDateError = "";

  function validateName(value: string): string {
    if (!value || value.trim().length === 0) {
      return "Name is required";
    }
    if (value.length > 100) {
      return "Name must be 100 characters or fewer";
    }
    return "";
  }

  function validateCustomDate(value: string): string {
    if (expirationOption !== "custom") return "";
    if (!value) {
      return "Please select an expiration date";
    }
    const selected = new Date(value);
    const now = new Date();
    now.setHours(0, 0, 0, 0);
    if (selected <= now) {
      return "Expiration date must be in the future";
    }
    return "";
  }

  function handleNameBlur() {
    nameError = validateName(name);
  }

  function handleCustomDateBlur() {
    customDateError = validateCustomDate(customDate);
  }

  function computeExpiresOn(): string | undefined {
    if (expirationOption === "none") {
      return undefined;
    }
    if (expirationOption === "custom") {
      if (!customDate) return undefined;
      return new Date(customDate).toISOString();
    }
    const days = parseInt(expirationOption, 10);
    const date = new Date();
    date.setDate(date.getDate() + days);
    return date.toISOString();
  }

  $: isFormValid =
    name.trim().length > 0 &&
    name.length <= 100 &&
    (expirationOption !== "custom" || (customDate && !validateCustomDate(customDate)));

  $: isSubmitting = $createToken.isPending;

  async function handleSubmit() {
    nameError = validateName(name);
    customDateError = validateCustomDate(customDate);

    if (nameError || customDateError) return;

    apiError = "";

    try {
      const expiresOn = computeExpiresOn();
      const result = await $createToken.mutateAsync({
        displayName: name.trim(),
        description: description.trim() || undefined,
        expiresOn,
      });

      createdToken = result.token ?? "";
      step = "display";
    } catch (err: any) {
      apiError =
        err?.response?.data?.message ||
        err?.message ||
        "Failed to create token. Please try again.";
    }
  }

  async function handleCopy() {
    if (!createdToken) return;
    try {
      await navigator.clipboard.writeText(createdToken);
      copied = true;
      setTimeout(() => {
        copied = false;
      }, 2000);
    } catch {
      // Fallback: select the text in the input
      const input = document.querySelector(
        "[data-token-display]",
      ) as HTMLInputElement;
      if (input) {
        input.select();
        document.execCommand("copy");
        copied = true;
        setTimeout(() => {
          copied = false;
        }, 2000);
      }
    }
  }

  function resetState() {
    step = "form";
    name = "";
    description = "";
    expirationOption = "90";
    customDate = "";
    createdToken = "";
    copied = false;
    apiError = "";
    nameError = "";
    customDateError = "";
  }

  function handleClose() {
    resetState();
    open = false;
    dispatch("close");
  }

  function getTomorrowDate(): string {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return tomorrow.toISOString().split("T")[0];
  }
</script>

<Dialog bind:open on:close={handleClose}>
  <svelte:fragment slot="title">
    {#if step === "form"}
      Create Personal Token
    {:else}
      Token Created Successfully
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="body">
    {#if step === "form"}
      <form
        on:submit|preventDefault={handleSubmit}
        class="flex flex-col gap-4"
      >
        <div class="flex flex-col gap-1.5">
          <label for="user-token-name" class="text-sm font-medium text-gray-700">
            Name <span class="text-red-500">*</span>
          </label>
          <input
            id="user-token-name"
            type="text"
            bind:value={name}
            on:blur={handleNameBlur}
            placeholder="e.g., CLI access, CI/CD pipeline"
            maxlength={100}
            class="w-full rounded-sm border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
            class:border-red-500={nameError}
          />
          {#if nameError}
            <p class="text-xs text-red-500">{nameError}</p>
          {/if}
        </div>

        <div class="flex flex-col gap-1.5">
          <label
            for="user-token-description"
            class="text-sm font-medium text-gray-700"
          >
            Description
            <span class="text-xs font-normal text-gray-400">(optional)</span>
          </label>
          <textarea
            id="user-token-description"
            bind:value={description}
            placeholder="What will this token be used for?"
            rows={2}
            class="w-full rounded-sm border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 resize-none"
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label
            for="user-token-expiration"
            class="text-sm font-medium text-gray-700"
          >
            Expiration
          </label>
          <select
            id="user-token-expiration"
            bind:value={expirationOption}
            class="w-full rounded-sm border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 bg-white"
          >
            {#each EXPIRATION_OPTIONS as option}
              <option value={option.value}>{option.label}</option>
            {/each}
          </select>
          <p class="text-xs text-gray-500">
            {#if expirationOption === "none"}
              This token will never expire. You can revoke it manually at any time.
            {:else if expirationOption === "custom"}
              Select a custom expiration date below.
            {:else}
              Token will expire {expirationOption} days from creation.
            {/if}
          </p>
        </div>

        {#if expirationOption === "custom"}
          <div class="flex flex-col gap-1.5">
            <label
              for="user-token-custom-date"
              class="text-sm font-medium text-gray-700"
            >
              Expiration Date <span class="text-red-500">*</span>
            </label>
            <input
              id="user-token-custom-date"
              type="date"
              bind:value={customDate}
              on:blur={handleCustomDateBlur}
              min={getTomorrowDate()}
              class="w-full rounded-sm border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
              class:border-red-500={customDateError}
            />
            {#if customDateError}
              <p class="text-xs text-red-500">{customDateError}</p>
            {/if}
          </div>
        {/if}

        <div class="rounded-sm border border-blue-200 bg-blue-50 px-3 py-2">
          <p class="text-xs text-blue-700">
            This token will have the same access permissions as your user account.
          </p>
        </div>

        {#if apiError}
          <div class="rounded-sm border border-red-200 bg-red-50 px-3 py-2">
            <p class="text-sm text-red-600">{apiError}</p>
          </div>
        {/if}
      </form>
    {:else}
      <div class="flex flex-col gap-4">
        <div class="flex items-center gap-2 rounded-sm border border-green-200 bg-green-50 px-3 py-2">
          <div class="flex size-5 items-center justify-center rounded-full bg-green-500 text-white">
            <Check size="12px" />
          </div>
          <p class="text-sm font-medium text-green-700">
            Personal token created successfully!
          </p>
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-gray-700">Your Token</label>
          <div class="flex gap-2">
            <input
              data-token-display
              type="text"
              value={createdToken}
              readonly
              class="w-full rounded-sm border border-gray-300 bg-gray-50 px-3 py-2 font-mono text-sm select-all"
            />
            <button
              type="button"
              on:click={handleCopy}
              class="flex shrink-0 items-center gap-1.5 rounded-sm border px-3 py-2 text-sm font-medium transition-colors
                {copied
                  ? 'border-green-300 bg-green-50 text-green-700'
                  : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'}"
            >
              {#if copied}
                <Check size="14px" />
                Copied
              {:else}
                <Copy size="14px" />
                Copy
              {/if}
            </button>
          </div>
        </div>

        <div class="rounded-sm border border-amber-200 bg-amber-50 px-3 py-3">
          <p class="text-sm font-medium text-amber-800">
            ⚠️ Make sure to copy your token now.
          </p>
          <p class="mt-1 text-xs text-amber-700">
            You won't be able to see it again! Store it securely — this token
            grants access to your Rill account.
          </p>
        </div>
      </div>
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="footer">
    {#if step === "form"}
      <div class="flex justify-end gap-2">
        <Button type="secondary" on:click={handleClose}>Cancel</Button>
        <Button
          type="primary"
          on:click={handleSubmit}
          disabled={!isFormValid || isSubmitting}
        >
          {#if isSubmitting}
            <span class="flex items-center gap-2">
              <span
                class="size-4 animate-spin rounded-full border-2 border-current border-t-transparent"
              />
              Creating...
            </span>
          {:else}
            Create Token
          {/if}
        </Button>
      </div>
    {:else}
      <div class="flex justify-end">
        <Button type="primary" on:click={handleClose}>Done</Button>
      </div>
    {/if}
  </svelte:fragment>
</Dialog>