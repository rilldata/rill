# Playwright Test Suite

Our Playwright test suite encompasses three main projects – `setup`, `e2e`, and `teardown`.

---

### `setup`

Playwright recommends that global setup takes place in a dedicated Playwright "project" ([docs](https://playwright.dev/docs/test-global-setup-teardown#option-1-project-dependencies)). This project:

1. Starts a fresh instance of Rill Cloud
2. Logs-in via the e2e-admin@rilldata.com user (which has been pre-populated in our Auth0 staging database)
3. Creates an organization named `e2e`
4. Deploys the OpenRTB project
5. Waits for data ingestion and asserts when the primary dashboard is ready-to-go

```bash
# Run with automatic teardown
npx playwright test --project=setup

# For development (without teardown)
npm run test:setup  # Shorthand for: E2E_NO_TEARDOWN=1 playwright test --project=setup
```

### `e2e`

The `e2e` project contains the main test suites.

```bash
# Run all tests (with setup and teardown)
npx playwright test --project=e2e

# Run a specific test (with setup and teardown)
npx playwright test --project=e2e -g 'My test'

# For development (without setup or teardown)
npm run test:dev -- -g 'My test'  # Shorthand for: E2E_NO_SETUP_OR_TEARDOWN=1 playwright test --project=e2e
```

### `teardown`

Handles cleanup:
1. Deletes the `e2e` organization
2. Shuts down all the services

```bash
npx playwright test --project=teardown
# Or use the shorthand: npm run test:teardown
```

### Development Workflow

When developing and debugging tests, follow this workflow to avoid repeatedly running setup and teardown:

1. Navigate to the `web-admin` directory.

2. Start the frontend in one terminal:
   ```bash
   npm run preview  # or npm run dev
   ```
   > **Note:** Use `npm run dev` for UI changes to avoid port conflicts with Playwright, which would otherwise build and run the UI on port 3000.

3. Set up your test environment (once per session):
   ```bash
   npm run test:setup
   ```

4. Run your tests as needed:
   ```bash
   npm run test:dev -- -g 'My test'
   ```

5. Clean up when finished:
   ```bash
   npm run test:teardown
   ```
   > **Important:** Always run teardown before using `rill devtool start cloud` to prevent port conflicts.