// SPDX-License-Identifier: AGPL-3.0-or-later

import { reconcileSuwayomiSettings } from '../domains/suwayomi-settings/application'

// Reconcile the Suwayomi settings Uchiyomi depends on at boot (ADR-0012):
// ENFORCE invariants (auto-download off) and SEED_IF_EMPTY the default repo.
// Non-fatal: a slow/unreachable Suwayomi must not block app startup.
export default defineNitroPlugin(async () => {
  const config = useRuntimeConfig()
  if (process.env.NODE_ENV === 'test' || config.suwayomiSettingsReconcile === false) {
    return
  }

  try {
    const report = await reconcileSuwayomiSettings.execute()
    // Only log when something actually changed; a healthy steady-state boot stays quiet.
    if (report.changed) {
      console.warn('[suwayomi-settings] reconciled:', report.changes.join('; '))
    }
  } catch (error) {
    console.warn('[suwayomi-settings] reconcile failed (will retry next boot):', error)
  }
})
