import { ConnectError } from "@connectrpc/connect"

/**
 * The message to show a Member for a failed request.
 *
 * Connect puts the code in front of the message ("invalid_argument: …"), which is for
 * the developer, not the person reading it. DESIGN.md §11 wants the plain sentence the
 * server wrote.
 */
export function messageFrom(error: unknown, fallback = "Something went wrong reaching the server."): string {
  if (error instanceof ConnectError) {
    return error.rawMessage
  }
  if (error instanceof Error) {
    return error.message
  }
  // The one string here that a Member reads; everything else comes from the server,
  // which writes in the Instance's own words.
  return fallback
}
