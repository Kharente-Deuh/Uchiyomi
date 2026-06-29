// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { ClientError } from 'graphql-request'
import { describe, expect, it } from 'vitest'
import { classifySuwayomiError, SuwayomiError } from '../../server/utils/suwayomi/errors'

describe('classifySuwayomiError', () => {
  it('passes a SuwayomiError through unchanged', () => {
    const original = new SuwayomiError('timeout', 'slow')
    expect(classifySuwayomiError(original)).toBe(original)
  })

  it('maps a graphql-request ClientError to kind "graphql"', () => {
    const err = new ClientError(
      { errors: [{ message: 'boom' }], status: 200, headers: new Headers() },
      { query: 'query {}' },
    )
    const result = classifySuwayomiError(err)
    expect(result).toBeInstanceOf(SuwayomiError)
    expect(result.kind).toBe('graphql')
  })

  it('maps any other error to kind "transport"', () => {
    const result = classifySuwayomiError(new Error('ECONNREFUSED'))
    expect(result.kind).toBe('transport')
    expect(result.cause).toBeInstanceOf(Error)
  })
})
