import { get } from "svelte/store";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { goto } from "$app/navigation";
import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";
import { getMetricsResolverQueryToUrlMapperStore } from "@rilldata/web-common/features/chat/core/messages/text/citation-url-mapper.ts";

/**
 * Adds a click handler to the given node, that intercepts clicks to links and uses svelte's goto.
 * Links added directly to the node are not intercepted by svelte so we need to manually intercept them.
 * Also adds a check to make sure url is actually a local link before using goto.
 */
export function enhanceCitationLinks(
  node: HTMLElement,
  conversation: Conversation,
) {
  const isEmbedded = EmbedStore.isEmbedded();
  const mapperStore = getMetricsResolverQueryToUrlMapperStore(conversation);

  function handleClick(e: MouseEvent) {
    if (!e.target || !(e.target instanceof HTMLElement)) return; // typesafety
    if (e.metaKey || e.ctrlKey) {
      if (isEmbedded) {
        // Prevent cmd/ctrl+click in embedded context. The iframe href used for embedding cannot be opened as is.
        e.preventDefault();
      }
      return;
    }

    const href = e.target.getAttribute("href") ?? "";
    const parsedUrl = URL.parse(href);

    const mapper = get(mapperStore).data;
    const mappedHref = parsedUrl && mapper ? mapper(parsedUrl) : href;

    const isLocalLink = window.location.origin === parsedUrl?.origin;
    const isPartialLink = mappedHref.startsWith("/");
    const shouldUseGoto = isLocalLink || isPartialLink;
    if (!shouldUseGoto) return;

    e.preventDefault();
    void goto(mappedHref);
  }

  node.addEventListener("click", handleClick);
  return {
    destroy() {
      node.removeEventListener("click", handleClick);
    },
  };
}
