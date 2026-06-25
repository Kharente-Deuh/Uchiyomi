// SPDX-License-Identifier: AGPL-3.0-or-later
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toPreferenceDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor || !actor.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'Missing id' })
  }

  const prefs = await extensionsService().listSourcePreferences({ sourceId: id })

  return { preferences: prefs.map(p => toPreferenceDto(p)) }
})
