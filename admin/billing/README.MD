# Billing Guide

note: If you are here to know how to bypass trial and restrictive quotas, directly jump to [Avoiding trial and restrictive quotas](#avoiding-trial-and-restrictive-quotas). 

For self serve billing, we use two external systems Orb and Stripe. Orb is used for managing the subscription and Stripe is used for payment processing.

## Orb
Orb is used to manage the subscription lifecycle, issue invoices as per the subscribed plan and auto-collect payment using Stripe for the issued invoices. It also supports usage based billing so currently we also report project duckdb size to Orb, and it can be used for billing.

### Subscription
A subscription is created when a user subscribes to a plan. It contains information about the plan, start date, end date, billing cycle etc.

### Plan
A plan is an entity that defines the pricing and quotas to be used like number of projects etc. We mainly have 3 plans - Starter, Teams and Enterprise. Starter is the default plan which has 30 days of free trial and limited quotas. Teams is paid plan that user can upgrade to after adding payment details in Stripe through our billing portal. Enterprise are custom plans manually assigned.

### Webhooks
Only invoice related events are handled to make sure payments are collected without any issues.

## Stripe
Stripe is used for payment processing. Customer add payment method and valid billable address in Stripe through our billing portal. Orb uses this information to auto-collect payment for the issued invoices.

### Payment Method
A payment method is a card that is added by the user in Stripe. It is used to pay for the invoices issued by Orb.

### Billable Address
A billable address is the address of the user used for tax purposes by Orb to issue invoices.

### Webhooks
Payment method and customer update related events are handled to make sure valid payment method and billable address are present.

## Implementation
Each org is linked to an Orb customer and Stripe customer using `billing_customer_id` and `payment_customer_id` fields in database. 

On creation of an org, a customer is created in Orb with `org_id` as the external customer id and a customer is created in Stripe, customer id returned from Stripe is updated in Orb customer as payment provider id.

### Billing Issues
Each org can have various billing issues like never-subscribed, on-trial, trial ended, payment method missing, no billable address, payment failed, subscription cancelled etc. Each issue has relevant metadata and much of the billing logic is based on these issues.

### River Jobs
River is used to run persistent background jobs for initializing billing for orgs, processing webhook events and running trial related checks and sending emails. When running through devtool river UI is available at `http://localhost:7070` and can be used to see the status of the jobs.

## Configuration
Following are the billing related configuration:
```
RILL_ADMIN_ORB_API_KEY=
RILL_ADMIN_STRIPE_API_KEY=
RILL_ADMIN_STRIPE_WEBHOOK_SECRET=
RILL_ADMIN_ORB_WEBHOOK_SECRET=
```
Both Orb and Stripe provides test-mode for development. Depending on the `...API_KEY` admin service will connect to the test or prod env.

Cloud .env has the test-mode values for first three and devtool starts stripe cli to forward Stripe events to the admin service. `RILL_ADMIN_ORB_WEBHOOK_SECRET` will be commented out as there is no automated way to set up utility to listed Orb events. Orb events are anyways used for handling rare scenarios.

### Manual Orb Webhook setup (optional)
If required, one can use ngrok cli to start a forwarder like this `ngrok http --domain=<ngrok-provided-ip> 8080`. Visit https://app.withorb.com/webhooks to add an endpoint with URL `https://<ngrok-provided-ip>/billing/webhook` and configure the signing secret as `RILL_ADMIN_ORB_WEBHOOK_SECRET`. 

## Usage
On new org creation there is no subscription created but few billing issues will be raised which can be viewed using `rill billing list-issues`.

### Trial plan
On first project deployment, org will automatically moved to trial plan having 30 days and after grace period of 9 more days, all projects will be hibernated until upgraded to a paid plan. Trial plan will have limited quotas for projects, slots etc.

#### Num Trial quota
By default, each user across lifetime is restricted to 2 trials. This is **not** checked for superusers and if needed can be increased for any user using `rill sudo quota set --user <email> --trial-orgs <num>`

### Avoiding trial and restrictive quotas
Immediately after org creation or at any time in the future, superusers can use `rill sudo org set-internal-plan` to set the org to internal plan which will have unlimited quotas and no trial.

If creating org through cli you can just use `rill org create <org-name> && rill sudo org set-internal-plan`.

### Continuing with Trial plan and upgrade
To continue with trial plan, user can upgrade to Teams plan after entering payment details and address in billing portal. In test-mode, use any of the various test cards listed [here](https://docs.stripe.com/testing#cards)

After fixing the issues, user can upgrade to Teams plan using cli as well `rill billing subscription edit --plan Teams`.

### Subscription cancel and renewal
User can cancel the subscription using `rill billing subscription cancel` and renew it using `rill billing subscription renew --plan <plan-name>`.

### Org deletion
On org deletion by `rill org delete` all the billing related data will be cleared and org will be removed from Orb and Stripe.

### Subscription related commands
```bash
rill billing subscription list
rill billing subscription edit --plan <plan-name>
rill billing subscription cancel
rill billing subscription renew --plan <plan-name>
```
`renew` can only be used if subscription is cancelled. `edit` can only be used to change plan and not renewal.

## Advanced use cases
Generally following are not allowed
- plan downgrade (meaning moving to a plan having less quotas than org's current quotas)
- moving to a paid plan without entering payment details and address or when there is an overdue payment
- moving to an Orb plan which does not have `public: true` set in metadata. Only Starter and Teams are public plans

but can be overridden using `rill billing subscription edit --plan <plan-name> --force` when editing plan or `rill billing subscription renew --plan <plan-name> --force` when renewing.

### Extending Trial
If needed, trial can be extended using `rill sudo billing extend-trial --org <org> --days <num-days>`.

### Deleting billing issues
If needed, billing issues can be deleted using `rill sudo billing delete-issue --org <org>`.
