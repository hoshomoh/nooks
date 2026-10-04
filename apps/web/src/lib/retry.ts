import { Code, ConnectError } from "@connectrpc/connect"

/** How many times a failed read is asked again before the failure is shown. */
const RETRIES = 2

/**
 * The answers that will be the same however many times the question is asked.
 *
 * Something deleted stays deleted, a List unshared stays out of reach, and a request the
 * Instance refused as malformed is refused again. Asking twice more only made whoever was
 * looking wait several seconds for an answer they already had.
 */
const FINAL: ReadonlySet<Code> = new Set([
  Code.NotFound,
  Code.PermissionDenied,
  Code.Unauthenticated,
  Code.InvalidArgument,
  Code.FailedPrecondition,
  Code.AlreadyExists,
  Code.OutOfRange,
  Code.Unimplemented,
])

/**
 * worthAskingAgain decides whether a failed read is retried.
 *
 * A home server can be slow, asleep, or busy for a moment, and those are worth a second
 * try. An answer is not.
 */
export function worthAskingAgain(failures: number, error: unknown): boolean {
  if (failures >= RETRIES) {
    return false
  }
  return !(error instanceof ConnectError && FINAL.has(error.code))
}
