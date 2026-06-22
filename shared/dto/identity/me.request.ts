// SPDX-License-Identifier: AGPL-3.0-or-later
// Request wire contracts for the self-service account routes (/api/auth/me*).
export interface UpdateMeRequestDto {
  displayName: string
}

export interface ChangePasswordRequestDto {
  currentPassword: string
  newPassword: string
  logoutOtherDevices?: boolean
}
