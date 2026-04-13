<script lang="ts">
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import OnboardingWorkspace from "@rilldata/web-common/features/onboarding/OnboardingWorkspace.svelte";
</script>

<div class="flex size-full overflow-hidden">
  <div class="flex size-full overflow-hidden bg-surface-subtle">
    <OnboardingWorkspace />
  </div>
  <DeveloperChat />
</div>
