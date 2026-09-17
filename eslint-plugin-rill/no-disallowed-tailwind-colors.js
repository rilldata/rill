/**
 * ESLint rule to disallow certain Tailwind text color classes.
 * Disallows: text-gray-*, text-neutral-*, text-slate-*, text-stone-*, text-zinc-*
 */

const DISALLOWED_TEXT_PATTERN =
  /\btext-(gray|neutral|slate|stone|zinc)-\d{1,3}\b/g;
const TEXT_ERROR_MESSAGE =
  'Disallowed Tailwind text color class: "{{ className }}". Use semantic color classes instead.';

const DISALLOWED_BACKGROUND_PATTERN =
  /\bbg-(gray|neutral|slate|stone|zinc)-\d{1,3}\b/g;
const DISALLOWED_BACKGROUND_CLASSES = /\bbg-(white|black)\b/g;
const BACKGROUND_ERROR_MESSAGE =
  'Disallowed Tailwind background color class: "{{ className }}". Use semantic color classes instead.';

const DISALLOWED_CLASSES_PATTERNS = [
  [DISALLOWED_TEXT_PATTERN, TEXT_ERROR_MESSAGE],

  [DISALLOWED_BACKGROUND_PATTERN, BACKGROUND_ERROR_MESSAGE],
  [DISALLOWED_BACKGROUND_CLASSES, BACKGROUND_ERROR_MESSAGE],
];

function reportAllMatches(value, context, node) {
  if (typeof value !== "string") return;

  for (const [pattern, errorMessage] of DISALLOWED_CLASSES_PATTERNS) {
    for (const match of value.matchAll(pattern)) {
      context.report({
        node,
        message: errorMessage,
        data: { className: match[0] },
      });
    }
  }
}

export default {
  meta: {
    type: "problem",
    docs: {
      description:
        "Disallow non-semantic Tailwind text/background color classes (gray, neutral, slate, stone, zinc)",
    },
    schema: [],
  },
  create(context) {
    const sourceCode = context.sourceCode ?? context.getSourceCode();

    return {
      // Check Svelte HTML attributes (class="..." and className="...")
      SvelteAttribute(node) {
        if (node.key?.name === "class" || node.key?.name === "className") {
          for (const valueNode of node.value) {
            if (valueNode.type === "SvelteLiteral") {
              reportAllMatches(valueNode.value, context, valueNode);
            }
          }
        }
      },
      // Check Svelte shorthand class directives (class:text-gray-500)
      SvelteDirective(node) {
        if (node.kind === "Class" && node.key?.name) {
          const className = node.key.name.name || node.key.name;
          reportAllMatches(className, context, node);
        }
      },
      // Check Svelte <style> blocks
      SvelteStyleElement(node) {
        const styleText = sourceCode.getText(node);
        const nodeStart = node.range[0];

        for (const [pattern, errorMessage] of DISALLOWED_CLASSES_PATTERNS) {
          for (const match of styleText.matchAll(pattern)) {
            const matchStart = nodeStart + match.index;
            const matchEnd = matchStart + match[0].length;

            context.report({
              loc: {
                start: sourceCode.getLocFromIndex(matchStart),
                end: sourceCode.getLocFromIndex(matchEnd),
              },
              message: errorMessage,
              data: { className: match[0] },
            });
          }
        }
      },
    };
  },
};
