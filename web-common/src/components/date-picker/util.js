export function findNestedMonthItem(monthItem) {
  const children = monthItem.parentNode.childNodes;
  for (let i = 0; i < children.length; i = i + 1) {
    const curNode = children.item(i);
    if (curNode === monthItem) {
      return i;
    }
  }
  return 0;
}
