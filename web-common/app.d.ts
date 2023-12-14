/**
 * Any custon action that is used on any svelteHTML element
 * must be declared here to enable TypeScript to
 * recognize that the action is allowed on that
 * svelteHtml element type.
 */

declare namespace svelteHTML {
  interface HTMLAttributes {
    // Used for copy action `shift-click-actions.ts`
    "on:shift-click"?: (event: CustomEvent) => void;
    "on:command-click"?: (event: CustomEvent) => void;
  }

  interface SVGAttributes {
    "on:scrub"?: (event: CustomEvent) => void;
    "on:scrub-start"?: (event: CustomEvent) => void;
    "on:scrub-end"?: (event: CustomEvent) => void;
    "on:scrub-move"?: (event: CustomEvent) => void;
    "on:shift-click"?: (event: CustomEvent) => void;
    "on:scrolling"?: (event: CustomEvent) => void;
  }
}
