// SPDX-License-Identifier: AGPL-3.0-or-later
import { userRepository } from '../../utils/identity'

export default defineEventHandler(async () => {
  return { required: (await userRepository.countUsers()) === 0 }
})
