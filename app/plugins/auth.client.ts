// SPDX-License-Identifier: AGPL-3.0-or-later

export default defineNuxtPlugin(async () => {
  const { getSetupStatus, fetchMe } = useAuth()

  // Detect first run before anything else: the guard reads needsAdmin from the store.
  const status = await getSetupStatus()

  // Only probe the session when an admin already exists.
  if (!status.success || !status.data.required) {
    await fetchMe()
  }
})
