import { ConnectError } from "@connectrpc/connect"

/**
 * The message to show a Member for a failed request.
 *
 * Connect puts the code in front of the message ("invalid_argument: …"), which is for
 * the developer, not the person reading it. DESIGN.md §11 wants the plain sentence the
 * server wrote.
 */
export function messageFrom(error: unknown): string {
  if (error instanceof ConnectError) {
    return error.rawMessage
  }
  if (error instanceof Error) {
    return error.message
  }
  return "Something went wrong reaching the server."
}
