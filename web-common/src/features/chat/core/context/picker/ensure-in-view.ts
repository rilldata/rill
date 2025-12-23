import type { InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";

export function ensureInView(node: HTMLElement, active: boolean) {
  if (active) {
    node.scrollIntoView({ block: "nearest" });
  }
  return {
    update(active: boolean) {
      if (active) {
        node.scrollIntoView({ block: "nearest" });
      }
    },
  };
}

export function ensureIsHighlighted(
  highlightManager: PickerOptionsHighlightManager,
) {
  return (node: Node, ctx: InlineContext) => {
    const onMouseEnter = () => highlightManager.highlightContext(ctx);
    node.addEventListener("mousemove", onMouseEnter);
    return {
      destroy() {
        node.removeEventListener("mousemove", onMouseEnter);
      },
    };
  };
}
