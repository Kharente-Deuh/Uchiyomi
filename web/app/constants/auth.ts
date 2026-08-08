// SPDX-License-Identifier: AGPL-3.0-or-later

export const ADMIN_ROUTE_GROUP = 'admin'
export const NOT_AUTHENTICATED_ROUTE_GROUP = 'not-authenticated'
export const AUTHENTICATED_ROUTE_GROUP = 'authenticated'

export type AuthRouteGroup
  = | typeof ADMIN_ROUTE_GROUP
    | typeof NOT_AUTHENTICATED_ROUTE_GROUP
    | typeof AUTHENTICATED_ROUTE_GROUP

declare module '#app' {
  interface PageMeta {
    authGroups?: AuthRouteGroup[]
  }
}
