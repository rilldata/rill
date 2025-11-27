<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createRuntimeServiceComplete,
    createRuntimeServiceUnpackEmpty,
    type V1Message,
  } from "@rilldata/web-common/runtime-client";
  import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";
  import { NEW_CONVERSATION_ID } from "@rilldata/web-common/features/chat/core/utils.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { get } from "svelte/store";
  import {
    MessageContentType,
    MessageType,
    ToolName,
  } from "@rilldata/web-common/features/chat/core/types.ts";
  import { goto } from "$app/navigation";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();
  const completeReq = createRuntimeServiceComplete();

  $: ({ instanceId } = $runtime);
  let prompt = "";
  let open = false;

  async function initProjectWithSampleData() {
    // await $unpackEmptyProject.mutateAsync({
    //   instanceId,
    //   data: {
    //     displayName: EMPTY_PROJECT_TITLE,
    //     force: true,
    //   },
    // });

    const conversation = new Conversation(instanceId, NEW_CONVERSATION_ID);
    conversation.draftMessage.set(
      `Generate a model for the following user prompt: ${prompt}`,
    );

    const messages = new Map<string, V1Message>();

    await conversation.sendMessage(
      { agent: ToolName.DEVELOPER_AGENT },
      {
        onMessage: (msg) => {
          messages.set(msg.id, msg);
          if (
            msg.type !== MessageType.RESULT ||
            msg.contentType === MessageContentType.ERROR
          )
            return;

          switch (msg.tool) {
            // Sometimes AI detects that model is already present.
            case ToolName.READ_FILE: {
              const callMsg = messages.get(msg.parentId);
              if (!callMsg) break;
              try {
                const content = JSON.parse(callMsg.contentData);
                eventBus.emit("notification", {
                  message: `Data already present at ${content.path}`,
                });
                open = false;
                void goto(`/files${content.path}`);
              } catch {
                // no-op
              }
              break;
            }

            case ToolName.WRITE_FILE: {
              const callMsg = messages.get(msg.parentId);
              if (!callMsg) break;
              try {
                const content = JSON.parse(callMsg.contentData);
                eventBus.emit("notification", {
                  message: `Data generated successfully at ${content.path}`,
                });
                open = false;
                void goto(`/files${content.path}`);
              } catch {
                // no-op
              }
              break;
            }
          }
        },
      },
    );
    await waitUntil(() => get(conversation.isStreaming));
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger asChild let:builder>
    <!--    <Button type="ghost" builders={[builder]} large>-->
    <!--      or generate sample data using AI-->
    <!--    </Button>-->
    <Button type="ghost" builders={[builder]} large>AI</Button>
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Generate sample data</Dialog.Title>
      <Dialog.Description>
        <div>What is the business context or domain of your data?</div>
      </Dialog.Description>
    </Dialog.Header>
    <Input id="sample-data" bind:value={prompt} />
    <Button
      type="primary"
      large
      loading={$completeReq.isPending}
      onClick={initProjectWithSampleData}
    >
      Generate
    </Button>
  </Dialog.Content>
</Dialog.Root>
