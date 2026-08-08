// SPDX-License-Identifier: AGPL-3.0-or-later
import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'

export const STATUS_PATH = '/status' satisfies RouteLocationRaw

export const SETUP_PATH = '/setup' satisfies RouteLocationRaw

function normalize(target: string): string {
  const separator = target.search(/[#?]/)
  const path = separator === -1 ? target : target.slice(0, separator)

  return path.toLowerCase().replace(/\/+$/, '')
}

export function isStatusPath(target: string): boolean {
  return normalize(target) === STATUS_PATH
}

export function statusRedirect(to: RouteLocationNormalized): RouteLocationRaw {
  return `${STATUS_PATH}?redirect=${encodeURIComponent(to.fullPath)}`
}
