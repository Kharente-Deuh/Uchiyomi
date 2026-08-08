// SPDX-License-Identifier: AGPL-3.0-or-later

export function initApi(endpoint: string = ''): ReturnType<typeof $fetch.create> {
  return $fetch.create({ baseURL: `/api${endpoint}` })
}
