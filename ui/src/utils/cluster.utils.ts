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

  // combine the 3 parts,, and truncate it at 40 characters, to keep it within the GCP limit
  return nameArray.join('-').slice(0, 40);
}
