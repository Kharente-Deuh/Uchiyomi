// SPDX-License-Identifier: AGPL-3.0-or-later
import type { H3Event } from 'h3'
import type { ExtensionModel } from '~~/server/domains/extensions/extension.domain'
import type { UserModel } from '~~/server/domains/identity/users/user.domain'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'

interface RequireAuthUserOpts {
  // Route requires the extension-management capability (403 otherwise).
  mustBeAbleToManage?: boolean
}

interface RequireExtensionOpts {
  installationStatus?: 'installed' | 'not-installed'
  byPassUpdateCheck?: boolean
  // When set, a user who canManageExtensions skips the installationStatus gate
  // entirely. Used by routes (e.g. the icon proxy) where managers legitimately
  // browse not-yet-installed extensions while regular users are restricted to
  // the installed set.
  adminBypassesInstallCheck?: boolean
}

type ExtensionGuardOpts = RequireAuthUserOpts & RequireExtensionOpts

// Cheap actor authorization — no I/O. Separated from the extension load so routes
// that validate a body/query can run authz (401/403) → input validation (400) →
// resource existence (404) in that order: a non-authorized caller never reaches,
// and never learns about, input validation or resource existence.
export function requireAuthUser(event: H3Event, opts?: RequireAuthUserOpts): UserModel {
  const { authUser } = event.context
  if (!authUser) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  if (opts?.mustBeAbleToManage && !authUser.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  return authUser
}

// Loads the extension from Suwayomi (the source of truth for installed / update /
// nsfw state) and applies the install + NSFW visibility gates. Hits the network,
// so only call it once cheap authz (requireAuthUser) and input validation have
// passed. Use it directly for routes with a body/query; the composed
// extensionGuard below is the shorthand for routes with neither.
export async function requireExtension(event: H3Event, authUser: UserModel, opts?: RequireExtensionOpts): Promise<ExtensionModel> {
  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const extension = await extensionsService().getExtensionByPkgName({ pkgName })
  if (!extension) {
    throw createError({ statusCode: 404, statusMessage: 'Extension not found' })
  }

  const enforceInstallCheck = !(opts?.adminBypassesInstallCheck && authUser.canManageExtensions)

  if (enforceInstallCheck && opts?.installationStatus === 'installed') {
    if (!extension.isInstalled) {
      throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
    }

    if (!opts?.byPassUpdateCheck && extension.hasUpdate) {
      throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
    }
  }

  if (enforceInstallCheck && opts?.installationStatus === 'not-installed' && extension.isInstalled) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  if (!authUser.allowNsfw && extension.isNsfw) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  return extension
}

// Shorthand for routes with no request body/query to validate: authorize the
// actor, then load and gate the extension in one call.
export async function extensionGuard(event: H3Event, opts?: ExtensionGuardOpts): Promise<{ authUser: UserModel, extension: ExtensionModel }> {
  const authUser = requireAuthUser(event, opts)
  const extension = await requireExtension(event, authUser, opts)

  return { authUser, extension }
}
