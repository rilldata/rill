<!-- 
  This layout wraps the chat page to provide proper height constraints
  and prevent double scrolling issues
-->

<div class="chat-page-wrapper">
  <slot />
</div>

<style lang="postcss">
  .chat-page-wrapper {
    /* Calculate available height after navigation elements */
    /* Approximate: TopNav (64px) + ProjectTabs (56px) = 120px */
    height: calc(100vh - 120px);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: #ffffff;
  }

  /* Ensure proper flex behavior for the chat component */
  :global(.chat-page-wrapper > .chat-fullpage) {
    flex: 1;
    min-height: 0; /* Important for proper flex shrinking */
  }
</style>
