// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import { catalogueService } from '~~/server/domains/catalogue/application/catalogue.service'
import { toSourceFiltersDto } from '~~/server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'
import { sourceGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/source.guard'
import { authGuard } from '~~/server/domains/identity/auth/infrastructure/http/guards/auth.guard'

export default defineEventHandler(async (event): Promise<SourceFilterDto[]> => {
  // Overlay-only route: a visible source in the overlay already implies its
  // extension is installed — authenticate, then load+gate the source.
  const authUser = authGuard(event)
  const source = await sourceGuard(event, authUser)

  const filters = await catalogueService().getSourceFilters({ sourceId: source.id })

  return toSourceFiltersDto(filters)
})
