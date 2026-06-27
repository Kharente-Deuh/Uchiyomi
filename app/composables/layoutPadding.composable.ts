// SPDX-License-Identifier: AGPL-3.0-or-later
import { useLayoutStore } from '~/store/layout.store'

interface LayoutPaddingsComposable {
  bottomLayout: ComputedRef<string | undefined>
  leftLayout: ComputedRef<string | undefined>
}

export function useLayoutPadding(): LayoutPaddingsComposable {
  const { mobile } = useDisplay()
  const layoutStore = useLayoutStore()

  const bottomLayout = computed(() => {
    if (mobile.value) {
      return 'var(--bottom-navigation-height)'
    }

    if (layoutStore.paginationEnabled) {
      return 'var(--bottom-pagination-height)'
    }
  })

  const leftLayout = computed(() => {
    if (mobile.value) {
      return
    }

    if (layoutStore.navigationDrawerCompact) {
      return 'var(--navigation-drawer-compact-width)'
    }

    return 'var(--navigation-drawer-width)'
  })

  return {
    bottomLayout,
    leftLayout,
  }
}
