# DraggableList Component

A generic, reusable draggable list component that can be used to create lists where items can be reordered via drag and drop.

## Basic Usage

```svelte
<script>
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  
  let items = [
    { id: "item1", name: "First Item" },
    { id: "item2", name: "Second Item" },
    { id: "item3", name: "Third Item" }
  ];

  function handleReorder(event) {
    items = event.detail.items;
  }
</script>

<DraggableList
  {items}
  on:reorder={handleReorder}
>
  <div slot="item" let:item>
    <span>{item.name}</span>
  </div>
</DraggableList>
```

## Props

- `items` - Array of items to display. Each item must have an `id` property.
- `searchValue` - Current search text (bindable)
- `showSearch` - Whether to show the search input
- `minHeight` - Minimum height of the list container
- `maxHeight` - Maximum height of the list container

## Events

- `on:reorder` - Fired when items are reordered. Event detail contains `{ items, fromIndex, toIndex }`
- `on:item-click` - Fired when an item is clicked. Event detail contains `{ item, index }`

## Slots

### `item` (required)
The main item slot. Available props:
- `item` - The item data
- `index` - The item index
- `isDragItem` - Whether this item is currently being dragged

### `header` (optional)
Optional header content. Available props:
- `items` - The filtered items

### `footer` (optional)
Optional footer content. Available props:
- `items` - The filtered items

### `search` (optional)
Custom search input. Available props:
- `searchValue` - The current search value

### `empty` (optional)
Content to show when no items are available. Available props:
- `searchValue` - The current search value

## Advanced Example

```svelte
<script>
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  
  let searchValue = "";
  let items = [
    { id: "task1", title: "Complete project", priority: "high" },
    { id: "task2", title: "Review code", priority: "medium" },
    { id: "task3", title: "Write documentation", priority: "low" }
  ];

  function handleReorder(event) {
    items = event.detail.items;
  }

  function handleItemClick(event) {
    console.log("Clicked item:", event.detail.item);
  }
</script>

<DraggableList
  {items}
  bind:searchValue
  showSearch={true}
  minHeight="200px"
  maxHeight="500px"
  on:reorder={handleReorder}
  on:item-click={handleItemClick}
>
  <div slot="header">
    <h3>My Tasks</h3>
  </div>

  <div slot="item" let:item let:index>
    <div class="flex items-center gap-2">
      <span class="flex-1">{item.title}</span>
      <span class="text-xs px-2 py-1 rounded bg-gray-100">
        {item.priority}
      </span>
    </div>
  </div>

  <div slot="empty" let:searchValue>
    {searchValue ? `No tasks matching "${searchValue}"` : "No tasks available"}
  </div>
</DraggableList>
```

## Features

- **Drag and drop reordering**: Items can be dragged to reorder them
- **Search filtering**: Built-in search functionality
- **Customizable slots**: Flexible content customization
- **Accessibility**: Keyboard navigation support
- **Responsive**: Works on touch devices
- **Type safe**: Full TypeScript support

## Styling

The component uses Tailwind CSS classes and follows the existing design system. You can customize the appearance by:

1. Using the provided slots to add your own styling
2. Overriding the default styles in your component
3. Using CSS custom properties for theme customization 