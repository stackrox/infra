/**
 * Asserts that the value is defined, i.e. not null or undefined.
 *
 * @param testValue value to test
 * @param msg message for the error thrown if test fails
 */
export default function assertDefined<T>(
  testValue: T,
  msg?: string,
): asserts testValue is NonNullable<T> {
  if (testValue === undefined || testValue === null) {
    throw new Error(msg || 'Must not be a nullable value');
  }
}
