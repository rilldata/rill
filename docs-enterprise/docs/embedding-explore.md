---
title: "⚙️ Embedding Explore"
slug: "embedding-explore"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="White-labeling your dashboards within existing applications" />

The Explore dashboard can be embedded into you own web application using an iFrame, a standard HTML element, and authenticated using our built-in support for single-sign-on (SSO) login.

Security is accomplished with a one-way hash using the secret key provided in the Rill account [administration page for SSO support](https://dash.rilldata.com/admin/#/sso).
## Embedding the Dashboard

To embed the Explore Dashboard into your web application, you simply need to place a SSO-authenticated iFrame into a HTML body of your application. Shown below is an example of the iFrame element for embedding the Rill dashboard:

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Rill iFrame Sample</title>
</head>
<body>
<h1>Rill iFrame Sample</h1>
<iframe src="http://dash.rilldata.com/sso/v1/login? ?firstName=John&lastName=Doe&email=demo%40rilldata.com&timestamp=1404419294&companyId=16&securityPolicyIds=518 &verifyHash=<SOME VERIFY HASH>&path=explore"  height="800" width="1200" seamless="seamless"></iframe>
</body>
</html>
```

## Single-Sign-On Authentication for the iFrame Element

To enable seamless integration with your application, we've built-in a SSO authentication mechanism for establishing user identity within the iFrame element, without requiring the user to explicitly login into Rill.

Specifically, we require the iFrame src field to point to our SSO login end-point with the following URL parameters:

| Parameter | Required | Description |
|---|---|---|
| firstName | Yes | First name of the user's account in the dashboard. Ex Jane |
| lastName | Yes | Last name of the user's account in the dashboard. Ex Smith |
| email | Yes | Email address of the user's account in the dashboard. Ex jsmith@example.com |
| companyId | Yes | A identification number for your company issued by your Rill representative. Ex 123 |
| securityPolicyIds | Yes | Comma-separated list of security policy ids to assign the user. At least one ID is required. Security-policy IDs are found in the Rill dashboard. Ex. 518,519,520 |
| timestamp | Yes | The current timestamp value in seconds (GMT/UTC format). Ex. 1358035200 |
| verifyHash | Yes | The MD5 hash of the secret key and the parameters listed above. Ex. b6036eb9f947695c46c9f4aee11be0b9 |
| path | No | The path to the view that will be displayed once the embedded dashboard loads. Ex. explore |

:::caution Email Address must be URL Encoded
email addresses (and any other strings containing characters unsafe for URLs) must be made URL safe by encoding them appropriately. The script examples below show how these strings are URL encoded.
:::

:::info verifyHash is based on MD5
verifyHash is a [MD5](https://en.wikipedia.org/wiki/MD5) hex hash, generated based on the values of the other URL parameters and your account's SSO Secret Key.
:::

## Obtaining the SSO Secret Key

The **Secret Key** is the core authentication token to ensure a secure connection between your application and the Rill dashboard, and should be **kept confidential**.

Initially, you must contact your Rill representative to generate the first secret key.

If the key's privacy is compromised, you can always reset a new Secret Key:

  * Log into your dashboard with an administrator account.
  * Go to the admin screens and choose Single Sign On from the left-hand sidebar menu.
  * Click Reset secret key. This causes a new secret key to be generated.
  * Double-click the key in the Current secret key field to select it, then copy the key.

**Store the key in a safe place.** 
## Example

Expanding on our sample above, below is a complete example of embedding the Rill dashboard into any HTML webpage.

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Rill iFrame Sample</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/crypto-js/3.1.9-1/core.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/crypto-js/3.1.9-1/md5.js"></script>
</head>
<body>
<h1>Rill iFrame Sample</h1>
<div id="rill_dash"></div>

<script type="text/javascript">
var generateHash = function(email, firstName, lastName, companyId, securityPolicyIds, secretKey, timestamp) {
    var stringToHash = firstName + '|' + lastName + '|' + email + '|' +
        companyId + '|' + securityPolicyIds + '|' + timestamp + '|' + secretKey
    var verify = CryptoJS.MD5(stringToHash).toString();
    return verify;
}
$(document).ready(function() {
    var firstName = "Jane";
    var lastName = "Smith";
    var email = "jsmith@example.com";
    var companyId = '123';
    var securityPolicyIds = "518,519,520";
    var secretKey = "423d004c716839b4af16ef680cc742f2";
    var timestamp = Math.floor(Date.now() / 1000);
    var verifyHash = generateHash(email, firstName, lastName, companyId, securityPolicyIds, secretKey, timestamp);
    var sso_src = "https://sso.rilldata.com/sso/v1/login?" +
        'firstName=' + encodeURIComponent(firstName) + '&' +
        'lastName=' + encodeURIComponent(lastName) + '&' +
        'email=' + encodeURIComponent(email) + '&' +
        'companyId=' + companyId + '&' +
        'securityPolicyIds=' + securityPolicyIds + '&' +
        'timestamp=' + timestamp + '&' +
        'verifyHash=' + verifyHash + '&' +
        'path=explore';
      $("div#rill_dash").html("<iframe src='" + sso_src + "' height='800' width='1200' seamless='seamless'></iframe>");
});
</script>
  
  </body>
</html>
```

## Creating New Users
The current version of the Rill dashboard SSO framework (v1) supports creation of new user accounts.   If a user authenticated via SSO does not exist, a Rill account is created for the user based on the information provided in the authenticated iFrame.
Accounts must have unique email addresses. If the provided email address is already associated with a Rill account, it is assumed to belong to the connecting user.
## Working With Security Policies
Each iFrame URL must specify one or more security policies IDs to assign to the account specified in the URL.  Security Policies need to be created by the account administrator first, before they can be accessed in the iFrame.
In the SSO URL, you must list the **numerical IDs **of the policies, not their names. Use the Rill Data dashboard to find the IDs in the security-policies administration pages.
