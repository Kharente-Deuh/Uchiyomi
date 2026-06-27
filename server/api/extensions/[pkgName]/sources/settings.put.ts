// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toExtensionSettingsDto } from '~~/server/domains/extensions/infrastructure/transport/http/extension-http.presenter'

// Discriminated on `type` so zod narrows the value field per branch. The body is the
// read-model echo, so the Suwayomi-guaranteed fields are required on their branches
// (entries/entryValues for list/multiSelect; booleanDefault for switch/checkbox); the
// per-pref current value stays optional (an unset value yields no write). `visible`
// defaults to true so callers may omit it.
const position = z.number().int().nonnegative()
const key = z.string().optional()
const visible = z.boolean().optional().default(true)

const PrefSchema = z.discriminatedUnion('type', [
  z.object({ position, type: z.literal('switch'), key, visible, booleanDefault: z.boolean(), booleanValue: z.boolean().optional() }),
  z.object({ position, type: z.literal('checkbox'), key, visible, booleanDefault: z.boolean(), booleanValue: z.boolean().optional() }),
  z.object({ position, type: z.literal('editText'), key, visible, textValue: z.string().optional() }),
  z.object({ position, type: z.literal('list'), key, visible, entries: z.array(z.string()), entryValues: z.array(z.string()), textValue: z.string().optional() }),
  z.object({ position, type: z.literal('multiSelect'), key, visible, entries: z.array(z.string()), entryValues: z.array(z.string()), multiValue: z.array(z.string()).optional() }),
])

const Body = z.object({
  common: z.array(PrefSchema),
  sources: z.array(z.object({
    id: z.string(),
    name: z.string().optional(),
    lang: z.string().optional(),
    preferences: z.array(PrefSchema),
  })),
})

export default defineEventHandler(async (event): Promise<ExtensionSettingsDto> => {
  const actor = event.context.authUser
  if (!actor?.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  const settings = await extensionsService().updateExtensionSettings({
    pkgName,
    settings: {
      pkgName,
      common: parsed.data.common,
      sources: parsed.data.sources.map(s => ({
        id: s.id,
        name: s.name ?? '',
        lang: s.lang ?? '',
        preferences: s.preferences,
      })),
    },
  })

  return toExtensionSettingsDto(settings)
})
