# Login and Sign up page template

This folder contains the log in and sign up which is used for Rill Cloud and Rill Enterprise. It's implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). Rollup is used for bundling the template.


### Project Details

The template is being used for both Rill Cloud and Rill Enterprise login. For cloud only two login options are activated, Google and windows. For Enterprise we additionally support Pingfed and Okta. The configuration of this is decided by the environment variables we pass in.

The forgot password/reset password feature is disabled for certain domains. These domains are mentioned through `VITE_DISABLE_FORGOT_PASS_DOMAINS`. 

## Running in development

1. Run `npm install -w web-auth`
2. Run `npm run dev -w web-auth`


## Building for production


1. Add `.env` file to the workspace with the contents

```
VITE_RILL_CLOUD_AUTH0_CLIENT_IDS="clientID1,clientID2,..."
VITE_OKTA_CONNECTION="<connection-name>"
VITE_PINGFED_CONNECTION="<connection-name>"
VITE_DISABLE_FORGOT_PASS_DOMAINS="domain1.com,domain2.com,..."
```
2. Run `npm run build -w web-auth`
3. Copy the contents of `bundle.html` and paste it to the Auth0 universal login page


**Note**: If static files such as fonts/favicon are changed, `template.html` should be updated with the new links. The static files are hosted on Rill CDN.