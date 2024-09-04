const appId = import.meta.env.RILL_UI_PUBLIC_PYLON_APP_ID as string;

// Right now both local and cloud return the same V1User from admin.
// Until we figure out a good place to put those common types (not the admin clients),
// this represents the user fields needed by pylon
export type UserLike = {
  email: string;
  displayName: string;
  photoUrl: string;
};

/**
 * Function implementation is copied from: https://docs.usepylon.com/chat/setup
 */
export function initPylonChat(user: UserLike) {
  window.pylon = {
    chat_settings: {
      app_id: appId,
      email: user.email,
      name: user.displayName,
      avatar_url: user.photoUrl,
    },
  };
}
