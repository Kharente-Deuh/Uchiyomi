// SPDX-License-Identifier: AGPL-3.0-or-later

import { useLayoutStore } from '~/stores/layout.store'

interface LayoutPaddingsComposable {
  bottomLayout: ComputedRef<string | undefined>
  leftLayout: ComputedRef<string | undefined>
  mainStyles: ComputedRef<Record<string, string>>
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

  const mainStyles = computed(() => {
    const styles: Record<string, string> = {}

    if (bottomLayout.value !== undefined) {
      styles['--v-layout-bottom'] = bottomLayout.value
    }

    if (leftLayout.value !== undefined) {
      styles['--v-layout-left'] = leftLayout.value
    }

    return styles
  })

  return {
    bottomLayout,
    leftLayout,
    mainStyles,
  }
}
