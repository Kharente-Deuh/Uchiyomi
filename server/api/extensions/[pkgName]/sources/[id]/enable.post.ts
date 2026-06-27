// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateSourceRequestDto } from '#shared/dto/extensions/extensions.request'
import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { requireAuthUser, requireExtension } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'
import { parseBody } from '~~/server/utils/request.util'
import { toSourceDto } from '../../../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

const BodySchema = z.object({ isEnabled: z.boolean() }) satisfies z.ZodType<UpdateSourceRequestDto>

export default defineEventHandler(async (event) => {
  const authUser = requireAuthUser(event, { mustBeAbleToManage: true })
  const body = await parseBody(event, BodySchema)
  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'Missing id' })
  }

  const extension = await requireExtension(event, authUser, { installationStatus: 'installed' })
  const { setSourceEnabled, getVisibleSource } = extensionsService()
  const src = await getVisibleSource({
    pkgName: extension.pkgName,
    sourceId: id,
    isAdmin: !!authUser.canManageExtensions,
    canSeeNsfw: !!authUser.allowNsfw && !!authUser.showNsfw,
  })
  if (!src) {
    throw createError({ statusCode: 404, statusMessage: 'Source not found' })
  }

  try {
    const source = await setSourceEnabled({ sourceId: id, isEnabled: body.isEnabled })

    return { source: toSourceDto(source) }
  } catch (err) {
    if (err instanceof Error && err.message === 'Source not found') {
      throw createError({ statusCode: 404, statusMessage: 'Source not found' })
    }

    throw err
  }
})
