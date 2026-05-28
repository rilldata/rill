Attachments are functions that run in an [effect]($effect) when an element is mounted to the DOM or when [state]($state) read inside the function updates.

Optionally, they can return a function that is called before the attachment re-runs, or after the element is later removed from the DOM.

> [!NOTE]
> Attachments are available in Svelte 5.29 and newer.

```svelte
<!--- file: App.svelte --->
<script>
	/** @type {import('svelte/attachments').Attachment} */
	function myAttachment(element) {
		console.log(element.nodeName); // 'DIV'

		return () => {
			console.log('cleaning up');
		};
	}
</script>

<div {@attach myAttachment}>...</div>
```

An element can have any number of attachments.

## Attachment factories

A useful pattern is for a function, such as `tooltip` in this example, to _return_ an attachment (demo:

```svelte
<!--- file: App.svelte --->
<script>
	import tippy from 'tippy.js';

	let content = $state('Hello!');

	/**
	 * @param {string} content
	 * @returns {import('svelte/attachments').Attachment}
	 */
	function tooltip(content) {
		return (element) => {
			const tooltip = tippy(element, { content });
			return tooltip.destroy;
		};
	}
</script>

<input bind:value={content} />

<button {@attach tooltip(content)}> Hover me </button>
```

Since the `tooltip(content)` expression runs inside an [effect]($effect), the attachment will be destroyed and recreated whenever `content` changes. The same thing would happen for any state read _inside_ the attachment function when it first runs. (If this isn't what you want, see [Controlling when attachments re-run](#Controlling-when-attachments-re-run).)

## Inline attachments

Attachments can also be created inline (demo:

```svelte
<!--- file: App.svelte --->
<canvas
	width={32}
	height={32}
	{@attach (canvas) => {
		const context = canvas.getContext('2d');

		$effect(() => {
			context.fillStyle = color;
			context.fillRect(0, 0, canvas.width, canvas.height);
		});
	}}
></canvas>
```

> [!NOTE]
> The nested effect runs whenever `color` changes, while the outer effect (where `canvas.getContext(...)` is called) only runs once, since it doesn't read any reactive state.

## Conditional attachments

Falsy values like `false` or `undefined` are treated as no attachment, enabling conditional usage:

```svelte
<div {@attach enabled && myAttachment}>...</div>
```

## Passing attachments to components

When used on a component, `{@attach ...}` will create a prop whose key is a [`Symbol`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Symbol). If the component then [spreads](/tutorial/svelte/spread-props) props onto an element, the element will receive those attachments.

This allows you to create _wrapper components_ that augment elements (demo:

```svelte
<!--- file: Button.svelte --->
<script>
	/** @type {import('svelte/elements').HTMLButtonAttributes} */
	let { children, ...props } = $props();
</script>

<!-- `props` includes attachments -->
<button {...props}>
	{@render children?.()}
</button>
```

```svelte
<!--- file: App.svelte --->
<script>
	import tippy from 'tippy.js';
	import Button from './Button.svelte';

	let content = $state('Hello!');

	/**
	 * @param {string} content
	 * @returns {import('svelte/attachments').Attachment}
	 */
	function tooltip(content) {
		return (element) => {
			const tooltip = tippy(element, { content });
			return tooltip.destroy;
		};
	}
</script>

<input bind:value={content} />

<Button {@attach tooltip(content)}>Hover me</Button>
```

## Controlling when attachments re-run

Attachments, unlike [actions](use), are fully reactive: `{@attach foo(bar)}` will re-run on changes to `foo` _or_ `bar` (or any state read inside `foo`):

```js
// @errors: 7006 2304 2552
function foo(bar) {
	return (node) => {
		veryExpensiveSetupWork(node);
		update(node, bar);
	};
}
```

In the rare case that this is a problem (for example, if `foo` does expensive and unavoidable setup work) consider passing the data inside a function and reading it in a child effect:

```js
// @errors: 7006 2304 2552
function foo(+++getBar+++) {
	return (node) => {
		veryExpensiveSetupWork(node);

+++		$effect(() => {
			update(node, getBar());
		});+++
	}
}
```

## Creating attachments programmatically

To add attachments to an object that will be spread onto a component or element, use [`createAttachmentKey`](svelte-attachments#createAttachmentKey).

## Converting actions to attachments

If you're using a library that only provides actions, you can convert them to attachments with [`fromAction`](svelte-attachments#fromAction), allowing you to (for example) use them with components.
