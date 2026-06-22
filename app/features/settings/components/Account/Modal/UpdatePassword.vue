<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import { object, string, ref as yupRef } from 'yup'
import { useForm } from '~/utils/forms/use-form'

const show = defineModel<boolean>({ required: true })
const { t } = useI18n()
const { loading, minPasswordLength } = useAuth()

function onSubmit(): void {}

const { field, handleSubmit, isValid, reset: resetForm } = useForm({
  initialValues: {
    currentPassword: '',
    newPassword: '',
    confirmNewPassword: '',
  },
  schema: object({
    currentPassword: string().required().label(t('settings.account.password.update.currentPassword')),
    newPassword: string()
      .required()
      .min(minPasswordLength.value)
      .notOneOf([yupRef('currentPassword')], t('settings.account.password.update.newPasswordSameAsCurrent'))
      .label(t('settings.account.password.update.newPassword')),
    confirmNewPassword: string()
      .required()
      .oneOf([yupRef('newPassword')], t('setup.passwordMismatch'))
      .label(t('settings.account.password.update.confirmNewPassword')),
  }),
  onSubmit,
})

watch(show, (value) => {
  if (!value) {
    loading.value = false
    resetForm()
  }
})
</script>

<template>
  <OrganismModal
    v-model="show"
    :title="$t('settings.account.password.change')"
    :loading
    :is-form-complete="isValid"
    prepend-icon="fa6-solid:key"
    @submit="handleSubmit"
  >
    <AtomInputPassword
      density="compact"
      v-bind="field('currentPassword').props"
      data-test="settings-account-password-current-password"
      :hide-details="false"
    />
    <AtomInputPassword
      density="compact"
      v-bind="field('newPassword').props"
      data-test="settings-account-password-new-password"
      :hide-details="false"
    />
    <AtomInputPassword
      density="compact"
      v-bind="field('confirmNewPassword').props"
      data-test="settings-account-password-confirm-new-password"
      :hide-details="false"
    />
  </OrganismModal>
</template>
