import namor from 'namor';

// eslint-disable-next-line import/prefer-default-export
export function generateClusterName(username = ''): string {
  // start with lowercase initials
  const userPart = username
    .toLocaleLowerCase()
    .split(' ')
    .map((namePiece) => namePiece[0])
    .join('');

  // extract the MM-DD part of the current date, ISO formatted
  const today = new Date();
  const datePart = today.toISOString().slice(5, 10);

  // finally, get a random 3-part string of real words
  const randomPart = namor.generate({ words: 3, saltLength: 0 });

  // prepare to combine, but filter any empty part (only one that should ever be empty is user string)
  const nameArray = [userPart, datePart, randomPart].filter(Boolean);

  // combine the 3 parts, truncate it at 28 characters, and remove a trailing '-', if any.
  // Truncation is needed to allow generating wildcard certificates for OpenShift using Let's Encrypt: since LE insists
  // on the domain forming the Common Name of the certificate, the string '*.apps.<name>.openshift.infra.rox.systems'
  // must not exceed 64 characters. This leaves a budget of 29 characters for the <name> portion.
  // RS-171 - 29 chars raises openshift error: "openshift_public_hostname must be 63 characters or less"
  return nameArray.join('-').slice(0, 28).replace(/-$/, '');
}
