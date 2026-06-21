// SPDX-License-Identifier: AGPL-3.0-or-later
// Hydrate the identity store from the session cookie before the route guard
// runs. `fetchMe` is a silent boot probe (401 -> store + cookie cleared).
export default defineNuxtPlugin(async () => {
  await useAuth().fetchMe()
})
