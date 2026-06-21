// SPDX-License-Identifier: AGPL-3.0-or-later
import type * as User from '../domains/identity/users/user.domain'

declare module 'h3' {
  interface H3EventContext {
    authUser: User.Model | null
  }
}
