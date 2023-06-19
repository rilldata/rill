# `admin/email`

This package contains logic and templates for sending transactional admin emails. 

## How templating works

We use [MJML](https://mjml.io/) to generate the email layout HTML. The MJML templates are found in the `templates` directory. 

To inject content into the email layout HTML, we use Go's built-in [`html/template`](https://pkg.go.dev/html/template) library.

We currently have just one template:

- `call_to_action.mjml` shows a title, body and button

## Adding/updating an MJML template

1. Add or edit the `.mjml` file in `templates` (we recommend using the [MJML online IDE](https://mjml.io/try-it-live/))
2. Run `./admin/email/templates/generate.sh`
3. Add a util function in `email.go` for loading and populating the template
