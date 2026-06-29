// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UserModel } from '../domains/identity/users/user.domain'

declare module 'h3' {
  interface H3EventContext {
    authUser: UserModel | null
  }
}
