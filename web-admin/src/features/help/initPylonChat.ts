import type { V1User } from "../../client";

export function initPylonChat(user: V1User) {
  window.pylon = {
    chat_settings: {
      // TODO: get the APP_ID from an environment variable
      app_id: "26a0fdd2-3bd3-41e2-82bc-1b35a444729f",
      email: user.email,
      name: user.displayName,
      avatar_url: user.photoUrl,
    },
  };
}
