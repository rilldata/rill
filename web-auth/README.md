
# Login and Sign up page template

  

This folder contains the log in and sign up which is used for Rill Cloud and Rill Enterprise. It's implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). Rollup is used for bundling the template.

## Project Details

The template is being used for both Rill Cloud and Rill Enterprise login. For cloud only two login options are activated, Google and windows. For Enterprise we additionally support Pingfed and Okta. The configuration of this is decided by the environment variables we pass in.

## Development

1. Add `.env` file to workspace
2. Run `npm install -w web-auth`
3. Run `npm run dev -w web-auth`

The development tooling uses SvelteKit. `+page.svelte` is being used to mount the `Auth` component and pass down the props/environment variables through it.  

The `configParams` props takes in `@@config@@` as an input. This is a variable which is replaced by the Auth0 runtime to an object containing tenant details. While in development we do not have the tenant object so login functionalities do not work. 

The project uses the following environment variables -
```
VITE_RILL_CLOUD_AUTH0_CLIENT_IDS="clientID1,clientID2,..."
VITE_OKTA_CONNECTION="<connection-name>"
VITE_PINGFED_CONNECTION="<connection-name>"
VITE_DISABLE_FORGOT_PASS_DOMAINS="domain1.com,domain2.com,..."
```
`VITE_RILL_CLOUD_AUTH0_CLIENT_IDS` is a comma separated list of Auth0 client IDs of application created for Rill Cloud.
`VITE_OKTA_CONNECTION` is the name of the connection set for Okta
`VITE_PINGFED_CONNECTION` is the name of the connection set for PingFed
`VITE_DISABLE_FORGOT_PASS_DOMAINS` is a comma separated list of domains for which reset password functionality has been blocked. This has been ported from the old sign-up template.

These environment variables can be found in 1Password. The document is named - **Rill Web Auth env**

While developing, to test the login features, deploy the generated template `bundle.html` (follow steps mentioned in building for production) to Auth0 staging. Verification can be done through Rill Cloud Staging and Dash staging.


## Building for production

1. Add `.env` file to the workspace
2. Run `npm run build -w web-auth`
3. Copy the contents of `bundle.html` 
4. Paste it in Auth0 login page which can be found at `https://manage.auth0.com/dashboard/us/<tenant-name>/login_page`

The build process uses Rollup extensively to package the template, inject JS inline and replacing the environment variables. 

**Note**: If static files such as fonts/favicon are changed, `template.html` should be updated with the new links. The static files are hosted on Rill CDN.

## Contributing

Before pushing a PR, deploy the latest build to staging so the reviewer can test it out. The same should be done on every new update/commit to the PR.

Once the PR has been approved, deploy it to Production Auth0.