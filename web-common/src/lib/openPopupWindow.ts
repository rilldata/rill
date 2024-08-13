const WindowWidth = 1024;
const WindowHeight = 600;

/**
 * Opens a popup window with {@link WindowWidth} and {@link WindowHeight}
 */
export function openPopupWindow(url: string, title: string) {
  const dualScreenLeft =
    window.screenLeft !== undefined ? window.screenLeft : window.screenX;
  const dualScreenTop =
    window.screenTop !== undefined ? window.screenTop : window.screenY;

  const width = window.innerWidth
    ? window.innerWidth
    : document.documentElement.clientWidth
      ? document.documentElement.clientWidth
      : screen.width;
  const height = window.innerHeight
    ? window.innerHeight
    : document.documentElement.clientHeight
      ? document.documentElement.clientHeight
      : screen.height;

  const systemZoom = width / window.screen.availWidth;
  const left = (width - WindowWidth) / 2 / systemZoom + dualScreenLeft;
  const top = (height - WindowHeight) / 2 / systemZoom + dualScreenTop;

  return window.open(
    url,
    title,
    `
      scrollbars=yes,
      width=${WindowWidth / systemZoom}, 
      height=${WindowHeight / systemZoom}, 
      top=${top}, 
      left=${left}
      `,
  );
}
