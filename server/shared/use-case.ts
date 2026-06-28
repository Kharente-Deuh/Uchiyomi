// SPDX-License-Identifier: AGPL-3.0-or-later

export interface IUseCase<Opts = any, Result = any> {
  execute: (request: Opts) => Promise<Result>
}
