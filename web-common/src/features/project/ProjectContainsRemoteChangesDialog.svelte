<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog/index.js";
  import CTACard from "@rilldata/web-common/components/calls-to-action/CTACard.svelte";

  export let open = false;
  export let loading = false;
  export let error: Error | null = null;
  export let onFetchAndMerge: () => void = () => {};
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild>
    <div class="hidden"></div>
  </Alert.Trigger>
  <Alert.Content class="min-w-[675px]" noCancel>
    <Alert.Header>
      <Alert.Title>Project updates available</Alert.Title>
      <Alert.Description>
        Your current project is out of date. There are newer changes available
        that aren’t in your current version.
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer>
      <div class="flex flex-col gap-y-2">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <CTACard
            title="Skip changes"
            description="Continue working with your current version. You can update later when you're ready."
            disabled={loading}
            ctaText="Skip changes"
            type="outlined"
            onClick={() => (open = false)}
          />
          <CTACard
            title="Fetch and merge changes"
            type="primary"
            {loading}
            disabled={loading}
            ctaText="Fetch and merge changes"
            onClick={onFetchAndMerge}
          >
            <svelte:fragment slot="description">
              <span class="font-medium text-primary-600">Recommended:</span>
              Get the latest version while preserving your work.
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
