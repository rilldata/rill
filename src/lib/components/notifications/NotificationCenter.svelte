<script>
    import store from ".";
    import Notification from "./Notification.svelte";
  </script>
  
  {#key $store.id}
    {#if $store.id}
      <Notification location=bottom {...$store.options}>
          {$store.message}
      </Notification>
    {/if}
  {/key}