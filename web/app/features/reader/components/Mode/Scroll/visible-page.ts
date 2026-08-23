// SPDX-License-Identifier: AGPL-3.0-or-later

export function recordIntersection(
  indices: Set<number>,
  index: number,
  isIntersecting: boolean,
): void {
  if (isIntersecting) {
    indices.add(index)
  } else {
    indices.delete(index)
  }
}

export function pageFromVisibleIndices(
  indices: Set<number>,
  opts: { restoredPage: number, restoring: boolean },
): number | undefined {
  if (indices.size === 0) {
    return undefined
  }

  const relevant = opts.restoring
    ? [...indices].filter(index => index >= opts.restoredPage)
    : [...indices]

  if (relevant.length === 0) {
    return undefined
  }

  return Math.min(...relevant)
}
