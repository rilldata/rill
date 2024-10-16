export function validateEmail(email) {
  const emailRegex =
    //eslint-disable-next-line
    /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

  return emailRegex.test(email);
}

export function getConnectionFromEmail(email, mapping) {
  const domain = email.split("@")[1];

  for (const connection in mapping) {
    const domainNames = mapping[connection];

    if (domainNames.includes(domain)) {
      return connection;
    }
  }

  return undefined; // No connection name found for the given email domain
}
