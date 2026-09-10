import { createClient } from "@connectrpc/connect"
import { AuthService, InstanceService } from "@nooks/api"

import { transport } from "./transport"

/** The generated clients, created once. */
export const instanceClient = createClient(InstanceService, transport)
export const authClient = createClient(AuthService, transport)
