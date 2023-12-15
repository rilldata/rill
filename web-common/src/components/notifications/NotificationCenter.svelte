<script lang="ts">
  import Notification from "./Notification.svelte";
  import PersistedLinkNotification from "./PersistedLinkNotification.svelte";
  import PersistedNotification from "./PersistedNotification.svelte";
  import notificationStore from "./notificationStore";
</script>

{#key $notificationStore.id}
  {#if $notificationStore.id}
    {#if $notificationStore?.options?.persisted}
      <PersistedNotification on:clear={() => notificationStore.clear()}>
        <div slot="title">{$notificationStore.message}</div>
        <div slot="body">{$notificationStore.detail}</div>
      </PersistedNotification>
    {:else if $notificationStore?.options?.persistedLink && $notificationStore.link}
      <PersistedLinkNotification
        message={$notificationStore.message}
        link={$notificationStore.link}
        on:clear={() => notificationStore.clear()}
      />
    {:else}
      <Notification location="bottom" {...$notificationStore.options}>
        {$notificationStore.message}
      </Notification>
    {/if}
  {/if}
{/key}
