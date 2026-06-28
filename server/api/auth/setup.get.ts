// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SetupStatusDto } from '#shared/dto/identity/auth.request'
import { userRepository } from '~~/server/domains/identity/users/application'

export default defineEventHandler(async (event): Promise<SetupStatusDto> => {
  const cfg = useRuntimeConfig(event).auth

  return {
    required: (await userRepository.countUsers()) === 0,
    minPasswordLength: cfg.minPasswordLength,
  }
})
