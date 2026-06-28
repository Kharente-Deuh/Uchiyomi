// SPDX-License-Identifier: AGPL-3.0-or-later

import type { H3Event } from 'h3'
import type { SourceModel } from '~~/prisma/generated/models'
import type { UserModel } from '~~/server/domains/identity/users/user.domain'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'

export async function sourceGuard(event: H3Event, authUser: UserModel, pkgName?: string): Promise<SourceModel> {
  const sourceId = getRouterParam(event, 'id')
  if (!sourceId) {
    throw createError({ statusCode: 400, statusMessage: 'Missing source id' })
  }

  if (!pkgName) {
    pkgName = getRouterParam(event, 'pkgName')
  }

  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const source = await extensionsService().getVisibleSource({
    pkgName,
    sourceId,
    isAdmin: !!authUser.canManageExtensions,
    canSeeNsfw: !!authUser.allowNsfw && !!authUser.showNsfw,
  })

  if (!source) {
    throw createError({ statusCode: 404, statusMessage: 'Source not found' })
  }

  return source
}
