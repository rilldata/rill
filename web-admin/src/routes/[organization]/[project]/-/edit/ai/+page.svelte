<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { onMount, tick } from "svelte";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "@rilldata/web-common/features/chat/core/conversation-manager";
  import ChatInput from "@rilldata/web-common/features/chat/core/input/ChatInput.svelte";
  import Messages from "@rilldata/web-common/features/chat/core/messages/Messages.svelte";
  import ConversationSidebar from "@rilldata/web-common/features/chat/layouts/fullpage/ConversationSidebar.svelte";
  import {
    conversationSidebarCollapsed,
    toggleConversationSidebar,
  } from "@rilldata/web-common/features/chat/layouts/fullpage/fullpage-store";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";

  const runtimeClient = useRuntimeClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: basePath = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit/ai`;

  $: conversationManager = getConversationManager(runtimeClient, {
    conversationState: "url",
    basePath: () => basePath,
  });

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  onMount(async () => {
    await tick();
    chatInputComponent?.focusInput();
  });

  beforeNavigate(({ to }) => {
    const isChatRoute = to?.url.pathname.includes("/-/edit/ai");
    if (!isChatRoute) {
      cleanupConversationManager(runtimeClient.instanceId);
    }
  });
</script>

<div class="chat-fullpage">
  <ConversationSidebar
    {conversationManager}
    {basePath}
    collapsed={$conversationSidebarCollapsed}
    onToggle={toggleConversationSidebar}
    onConversationClick={() => chatInputComponent?.focusInput()}
    onNewConversationClick={() => chatInputComponent?.focusInput()}
  />

  <div class="chat-main">
    <div class="chat-content">
      <div class="chat-messages-wrapper">
        <Messages
          {conversationManager}
          layout="fullpage"
          config={projectChat}
        />
      </div>
    </div>

    <div class="chat-input-section">
      <div class="chat-input-wrapper">
        <ChatInput
          {conversationManager}
          onSend={onMessageSend}
          bind:this={chatInputComponent}
          config={projectChat}
        />
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .chat-fullpage {
    display: flex;
    height: 100%;
    width: 100%;
    background: var(--surface);
  }
  .chat-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--surface);
  }
  .chat-content {
    flex: 1;
    overflow: hidden;
    background: var(--surface);
    display: flex;
    flex-direction: column;
  }
  .chat-messages-wrapper {
    flex: 1;
    overflow-y: auto;
    width: 100%;
    display: flex;
    flex-direction: column;
  }
  .chat-input-section {
    flex-shrink: 0;
    background: var(--surface);
    padding: 1rem;
    display: flex;
    justify-content: center;
  }
  .chat-input-wrapper {
    width: 100%;
    max-width: 48rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
</style>
