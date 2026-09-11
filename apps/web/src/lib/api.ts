import { createClient } from "@connectrpc/connect"
import {
  ActivityService,
  AuthService,
  InstanceService,
  ListService,
  MemberService,
  PublicService,
  RequestService,
  TokenService,
} from "@nooks/api"

import { transport } from "./transport"

/** The generated clients, created once. */
export const instanceClient = createClient(InstanceService, transport)
export const authClient = createClient(AuthService, transport)
export const listClient = createClient(ListService, transport)
export const memberClient = createClient(MemberService, transport)
export const activityClient = createClient(ActivityService, transport)
export const requestClient = createClient(RequestService, transport)
export const publicClient = createClient(PublicService, transport)
export const tokenClient = createClient(TokenService, transport)
