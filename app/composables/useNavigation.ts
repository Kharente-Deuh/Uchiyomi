// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ComputedRef } from 'vue'

export interface NavItem {
  key: string
  to: string
  icon: string
  labelKey: string
  /** Reserved for role-gated destinations (e.g. Admin, M6.2). */
  adminOnly?: boolean
}

const DESTINATIONS: NavItem[] = [
  { key: 'home', to: '/', icon: 'fa6-solid:house', labelKey: 'nav.home' },
  { key: 'settings', to: '/settings', icon: 'fa6-solid:gear', labelKey: 'nav.settings' },
]

export function useNavigation(): { items: ComputedRef<NavItem[]> } {
  const authStore = useAuthStore()
  const items = computed(() => DESTINATIONS.filter(item => !item.adminOnly || authStore.isAdmin))

  return { items }
}
