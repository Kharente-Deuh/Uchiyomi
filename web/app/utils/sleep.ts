// SPDX-License-Identifier: AGPL-3.0-or-later

export async function sleep(d: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, d))
}
