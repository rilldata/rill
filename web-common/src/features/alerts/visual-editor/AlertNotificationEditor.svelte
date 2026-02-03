<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";

  export let enableSlackNotification: boolean;
  export let slackChannels: string[];
  export let slackUsers: string[];
  export let enableEmailNotification: boolean;
  export let emailRecipients: string[];

  // These are unused now but kept for interface compatibility
  void enableSlackNotification;
  void enableEmailNotification;
</script>

<div class="flex flex-col gap-y-4">
  <!-- Slack Notifications -->
  <div class="flex flex-col gap-y-2">
    <div class="flex items-center justify-between">
      <Label class="text-sm font-medium text-fg-primary">
        Slack notifications
      </Label>
      <a
        href="https://docs.rilldata.com/guide/alerts#configuring-slack-targets"
        target="_blank"
        class="text-xs text-primary-600 hover:underline"
      >
        Learn more
      </a>
    </div>

    <div class="flex flex-col gap-y-3 pt-1">
      <MultiInput
        id="slackChannels"
        placeholder="# Enter a Slack channel name"
        description="Send alerts directly to these channels."
        contentClassName="relative"
        bind:values={slackChannels}
        singular="channel"
        plural="channels"
        preventFocus={true}
      />
      <MultiInput
        id="slackUsers"
        placeholder="Enter an email address"
        description="Alert users with direct messages in Slack."
        contentClassName="relative"
        bind:values={slackUsers}
        singular="user"
        plural="users"
        preventFocus={true}
      />
    </div>
  </div>

  <!-- Email Notifications -->
  <div class="flex flex-col gap-y-2">
    <Label class="text-sm font-medium text-fg-primary">
      Email notifications
    </Label>

    <div class="pt-1">
      <MultiInput
        id="emailRecipients"
        placeholder="Enter an email address"
        description="Recipients must have access to your project."
        contentClassName="relative"
        bind:values={emailRecipients}
        singular="email"
        plural="emails"
        preventFocus={true}
      />
    </div>
  </div>
</div>
