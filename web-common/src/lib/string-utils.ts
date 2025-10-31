export function snakeToCamel(snake: string): string {
  let camel = snake[0];
  for (let i = 1; i < snake.length; i++) {
    const isUnderscore = snake[i] === "_";
    if (!isUnderscore) {
      camel += snake[i];
      continue;
    }

    i++;
    if (i >= snake.length) break;

    camel += snake[i].toUpperCase();
  }
  return camel;
}

export function camelToSnake(camel: string): string {
  return camel.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
}
