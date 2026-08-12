// SPDX-License-Identifier: AGPL-3.0-or-later

export const ASURA_SORT_VALUES = ['popular', 'latest', 'rating', 'title', 'newest'] as const
export type AsuraSort = (typeof ASURA_SORT_VALUES)[number]

export const ASURA_SORT_ORDER_VALUES = ['asc', 'desc'] as const
export type AsuraSortOrder = (typeof ASURA_SORT_ORDER_VALUES)[number]

export const ASURA_STATUS_VALUES = ['ongoing', 'completed', 'hiatus', 'cancelled'] as const
export type AsuraStatus = (typeof ASURA_STATUS_VALUES)[number]

export const ASURA_TYPE_VALUES = ['manga', 'manhua', 'manhwa', 'mangatoon'] as const
export type AsuraType = (typeof ASURA_TYPE_VALUES)[number]
