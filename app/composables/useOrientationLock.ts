// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ComputedRef } from 'vue'
import { useDisplay } from 'vuetify'

/** Pure: lock portrait only on an enabled context where the API is available. */
export function shouldLockPortrait(enabled: boolean, isSupported: boolean): boolean {
  return enabled && isSupported
}

/** Pure: show the "rotate" overlay when enabled and the device is in landscape. */
export function shouldShowOrientationOverlay(enabled: boolean, orientation?: string): boolean {
  return enabled && (orientation?.startsWith('landscape') ?? false)
}

export function useOrientationLock(): { showOverlay: ComputedRef<boolean> } {
  const route = useRoute()
  const display = useDisplay()
  const { isSupported, orientation, lockOrientation } = useScreenOrientation()

  // Mobile only, unless the page opted out via `definePageMeta({ orientation: 'free' })`.
  const enabled = computed(() => display.mobile.value && route.meta.orientation !== 'free')

  watchEffect(() => {
    if (!import.meta.client) {
      return
    }

    if (shouldLockPortrait(enabled.value, isSupported.value)) {
      // Lock can reject when not fullscreen/installed — the overlay is the fallback.
      Promise.resolve(lockOrientation('portrait')).catch(() => {})
    }
    // NOTE (M5.2): when a route opts out (enabled -> false) we should also call
    // unlockOrientation() to release a previously-acquired lock; no opted-out
    // route exists yet, so this is currently inert.
  })

  const showOverlay = computed(() => shouldShowOrientationOverlay(enabled.value, orientation.value ?? undefined))

  return { showOverlay }
}
