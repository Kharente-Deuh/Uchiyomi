// SPDX-License-Identifier: AGPL-3.0-or-later

export const useLayoutStore = defineStore('layout', () => {
  const paginationEnabled = ref(false)
  const navigationDrawerCompact = ref(false)

  function setPaginationEnabled(isEnabled: boolean): void {
    paginationEnabled.value = isEnabled
  }

  function setNavigationDrawerCompact(isCompact: boolean): void {
    navigationDrawerCompact.value = isCompact
  }

  return {
    paginationEnabled,
    setPaginationEnabled,
    navigationDrawerCompact,
    setNavigationDrawerCompact,
  }
})
