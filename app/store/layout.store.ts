export const useLayoutStore = defineStore('layout', () => {
  const paginationEnabled = ref(false)
  const navigationDrawerCompact = ref(false)

  function setPaginationEnabled(value: boolean): void {
    paginationEnabled.value = value
  }

  function setNavigationDrawerCompact(value: boolean): void {
    navigationDrawerCompact.value = value
  }

  return {
    paginationEnabled,
    setPaginationEnabled,
    navigationDrawerCompact,
    setNavigationDrawerCompact,
  }
})
