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
  opts: { movingToPage: number | null },
): number | undefined {
  if (indices.size === 0) {
    return undefined
  }

  if (opts.movingToPage !== null) {
    return indices.has(opts.movingToPage) ? opts.movingToPage : undefined
  }

  return Math.min(...indices)
}
