const WindowWidth = 1024;
const WindowHeight = 600;

export class PopupWindow {
  private window: Window | null;
  private timer: ReturnType<typeof setInterval> | null;

  /**
   * Opens a popup window with {@link WindowWidth} and {@link WindowHeight}
   */
  public static open(url: string, title: string) {
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

  public openAndWaitForClose(url: string) {
    return new Promise<void>((resolve) => {
      try {
        // safeguard try catch
        this.window?.close();
      } catch {
        // no-op
      }
      if (this.timer) clearInterval(this.timer);
      this.window = PopupWindow.open(url, "popupWindow");

      // periodically check if the new window was closed
      this.timer = setInterval(() => {
        if (!this.window?.closed) return;
        clearInterval(this.timer as any);
        this.timer = null;
        this.window = null;
        resolve();
      }, 200);
    });
  }
}
