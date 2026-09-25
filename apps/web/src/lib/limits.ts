/**
 * The most a Member may write into each field.
 *
 * These are the Instance's numbers, not the app's: the server refuses anything longer,
 * and this is what lets a field say so while somebody is typing rather than after they
 * press save. `limits.test.ts` holds every one of them to the proto that declares it,
 * so the two cannot drift.
 *
 * Counted in characters rather than bytes, which is what the number means to whoever is
 * typing. The proto says max_len, which counts Unicode code points, and so does the
 * server.
 */
export const LIMITS = {
  itemLabel: 500,
  itemQuantity: 50,
  itemNote: 64_000,
  listName: 200,
  memberName: 100,
  /** The longest an address can be, from RFC 5321. Not a judgement. */
  memberEmail: 254,
  groupName: 200,
  tokenName: 200,
  instanceName: 200,
  joinMessage: 1_000,
} as const
