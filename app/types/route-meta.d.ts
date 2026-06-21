// SPDX-License-Identifier: AGPL-3.0-or-later
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    /** `'free'` opts a page out of the mobile portrait-lock (e.g. the reader). */
    orientation?: 'free'
  }
}
