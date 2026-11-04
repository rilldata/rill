// Client script of the embed host, shared by index.html and open.html.

const params = new URLSearchParams(window.location.search);

/** @returns {Promise<{ key: string, label: string }[]>} */
async function fetchUsers() {
  const resp = await fetch("/users.json");
  return resp.json();
}

// open.html: links to the host page as each user, keeping `rill_open` and any other params.
async function renderOpenPage() {
  const container = document.getElementById("open-users");
  if (!container) return;
  for (const user of await fetchUsers()) {
    const href = new URLSearchParams(params);
    href.set("user", user.key);
    const link = document.createElement("a");
    link.href = `/?${href.toString()}`;
    link.textContent = user.label;
    container.appendChild(link);
  }
}

// index.html: requests an iframe URL from the host's backend and renders it.
async function renderHostPage() {
  const usersEl = document.getElementById("users");
  const navigationEl = /** @type {HTMLInputElement} */ (
    document.getElementById("navigation")
  );
  const errorEl = /** @type {HTMLElement} */ (
    document.getElementById("embed-error")
  );
  const iframe = /** @type {HTMLIFrameElement} */ (
    document.getElementById("rill-frame")
  );
  if (!usersEl || !navigationEl || !errorEl || !iframe) return;

  let user = params.get("user") ?? "none";
  let navigation = params.get("navigation") !== "false";
  let iframeOrigin = "";

  // Relays the embed's postMessage traffic to the console, where tests record it.
  window.addEventListener("message", (event) => console.log(event.data));

  function syncUrl() {
    params.set("user", user);
    params.set("navigation", String(navigation));
    window.history.replaceState(null, "", `/?${params.toString()}`);
  }

  async function loadIframe() {
    errorEl.hidden = true;
    const resp = await fetch("/api/iframe", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        user,
        navigation,
        resource: params.get("resource"),
        open: params.get("rill_open"),
      }),
    });
    const data = await resp.json().catch(() => ({}));
    if (!resp.ok) {
      errorEl.textContent =
        data.message ?? `Request failed with status ${resp.status}`;
      errorEl.hidden = false;
      iframe.hidden = true;
      iframe.removeAttribute("src");
      return;
    }
    iframeOrigin = new URL(data.iframeSrc).origin;
    iframe.src = data.iframeSrc;
    iframe.hidden = false;
  }

  const userButtons = new Map();
  function renderSelectedUser() {
    for (const [key, button] of userButtons) {
      button.setAttribute("aria-pressed", String(key === user));
    }
  }
  for (const u of await fetchUsers()) {
    const button = document.createElement("button");
    button.type = "button";
    button.textContent = u.label;
    button.addEventListener("click", () => {
      user = u.key;
      renderSelectedUser();
      syncUrl();
      void loadIframe();
    });
    userButtons.set(u.key, button);
    usersEl.appendChild(button);
  }
  renderSelectedUser();

  navigationEl.checked = navigation;
  navigationEl.addEventListener("change", () => {
    navigation = navigationEl.checked;
    syncUrl();
    void loadIframe();
  });

  for (const button of document.querySelectorAll("button[data-page]")) {
    button.addEventListener("click", () => {
      iframe.contentWindow?.postMessage(
        {
          method: "navigate",
          params: {
            type: "page",
            name: /** @type {HTMLElement} */ (button).dataset.page,
          },
        },
        iframeOrigin,
      );
    });
  }

  await loadIframe();
}

void renderOpenPage();
void renderHostPage();
