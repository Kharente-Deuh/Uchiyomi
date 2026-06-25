export const useLayoutStore = defineStore('layout', () => {
  const paginationEnabled = ref(false)

  function setPaginationEnabled(value: boolean): void {
    paginationEnabled.value = value
  }

  return {
    paginationEnabled,
    setPaginationEnabled,
  }
})
