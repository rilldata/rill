<script lang="ts">
  import notificationStore from "./index";
  import Notification from "./Notification.svelte";
  import PersistedNotification from "./PersistedNotification.svelte";
</script>

{#key $notificationStore.id}
  {#if $notificationStore.id}
    {#if $notificationStore.options.persisted}
      <PersistedNotification on:clear={() => notificationStore.clear()}>
        <div slot="title">{$notificationStore.message}</div>
        <div slot="body">{$notificationStore.detail}</div>
      </PersistedNotification>
    {:else}
      <Notification location="bottom" {...$notificationStore.options}>
        {$notificationStore.message}
      </Notification>
    {/if}
  {/if}
{/key}
