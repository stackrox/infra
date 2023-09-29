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

  // ROX-15492
  //   - OCP 3.11 flavor allows us a maximum of up to 27 characters to fulfil: "openshift_public_hostname must be 63 characters or less".
  //   - OCP 4 demo: Let's Encrypt allows common names of up to 63 characters. With the format '*.apps.<name>.openshift.infra.rox.systems', 28 characters remain for the name.
  // Combine the 3 parts, truncate it at 27 characters, and remove a trailing '-', if any.
  return nameArray.join('-').slice(0, 27).replace(/-$/, '');
}
