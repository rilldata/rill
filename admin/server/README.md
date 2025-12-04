# admin/server

## GitHub Connection

### State Parameter in GitHub Authentication

When integrating with GitHub, installation request for our github app utilize a "state" parameter that is passed to GitHub and returned during the callback process. This mechanism helps maintain context across the OAuth flow.

#### For GitHub Installation Requests

The state parameter contains a JSON object with two key components:

1. **Repository Association**: `repo` field is used to store the specific repo associated with the original installation request from either the cli or UI. This is used to verify that the installation succeeded.

2. **UI Redirection**: `redirect` field is used to store the redirect url sent in the original installation request. This is mainly used by UI based deploy in rill developer to continue deploy process in the browser.

#### For GitHub Authentication-Only Requests

For authentication-only requests (without installation):
- The same two values are stored in the user's session
- These values are retrieved during the callback process
- This approach maintains context across the authentication flow