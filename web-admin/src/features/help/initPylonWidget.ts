const appId = import.meta.env.RILL_UI_PUBLIC_PYLON_APP_ID as string;

/**
 * Function implementation is copied from https://docs.usepylon.com/chat/setup
 */
export async function initPylonWidget() {
  const e = window;
  const t = document;
  const n = function () {
    n.e(arguments);
  };
  n.q = [];
  n.e = function (e) {
    n.q.push(e);
  };
  e.Pylon = n;
  const r = function () {
    const e = t.createElement("script");
    e.setAttribute("type", "text/javascript");
    e.setAttribute("async", "true");
    e.setAttribute("src", `https://widget.usepylon.com/widget/${appId}`);
    const n = t.getElementsByTagName("script")[0];
    n.parentNode.insertBefore(e, n);
  };
  if (t.readyState === "complete") {
    r();
  } else if (e.addEventListener) {
    e.addEventListener("load", r, false);
  }
}
