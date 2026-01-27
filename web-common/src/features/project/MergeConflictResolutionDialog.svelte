<script lang="ts">
  import type { ConnectError } from "@connectrpc/connect";
  import * as Alert from "@rilldata/web-common/components/alert-dialog/index.js";
  import CTACard from "@rilldata/web-common/components/calls-to-action/CTACard.svelte";
  import { AlertCircleIcon } from "lucide-svelte";

  export let open = false;
  export let loading = false;
  export let error: ConnectError | null = null;
  export let onUseLatestVersion: () => void = () => {};
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild>
    <div class="hidden"></div>
  </Alert.Trigger>
  <Alert.Content class="min-w-[675px]" noCancel>
    <Alert.Header>
      <Alert.Title>Merge conflicts detected</Alert.Title>
      <Alert.Description>
        <div>
          Your changes and the latest version have conflicting edits that cannot
          be automatically combined.
        </div>
        <div
          class="flex flex-row mt-4 p-4 gap-x-2 bg-yellow-50 border border-amber-100"
        >
          <AlertCircleIcon class="text-yellow-700 -mt-0.5" size={28} />
          <div class="flex flex-col gap-y-1">
            <h3 class="text-base font-medium text-fg-secondary">
              What are merge conflicts?
            </h3>
            <div class="text-sm text-fg-tertiary">
              Conflicts occur when same part of the file has been changed in
              different ways. You need to choose which version to keep.
            </div>
          </div>
        </div>
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer>
      <div class="flex flex-col gap-y-2">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <CTACard
            title="Keep my version"
            description="Not recommended: You will keep your changes, but may face deployment issues later."
            disabled={loading}
            ctaText="Keep my version"
            onClick={() => (open = false)}
          />
          <CTACard
            title="Use the latest changes"
            type="primary"
            {loading}
            disabled={loading}
            ctaText="Use the latest changes"
            onClick={onUseLatestVersion}
          >
            <svelte:fragment slot="description">
              <span class="font-medium text-primary-600">Recommended:</span>
              Your changes will be backed up before being replaced.
            </svelte:fragment>
          </CTACard>
        </div>
        {#if error}
          <div class="text-red-600">{error.message}</div>
        {/if}
      </div>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
