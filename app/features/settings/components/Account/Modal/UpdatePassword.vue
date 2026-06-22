<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import { object, string, ref as yupRef } from 'yup'
import { useForm } from '~/utils/forms/use-form'

const show = defineModel<boolean>({ required: true })
const { t } = useI18n()
const { loading, minPasswordLength, changePassword } = useAuth()
const toast = useToast()

// A plain ref (not a form field): the checkbox needs no validation, and the
// form's field-props bundle is shaped for text inputs, not VCheckbox.
const logoutOtherDevices = ref(false)

interface Form {
  currentPassword: string
  newPassword: string
  confirmNewPassword: string
}

async function onSubmit(values: Form): Promise<void> {
  const res = await changePassword({
    currentPassword: values.currentPassword,
    newPassword: values.newPassword,
    logoutOtherDevices: logoutOtherDevices.value,
  })

  if (res.success) {
    toast.success(t('settings.account.password.updated'))
    show.value = false

    return
  }

  toast.error(t('settings.account.password.updateError'))
}

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
    logoutOtherDevices.value = false
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

    <VCheckbox
      v-model="logoutOtherDevices"
      :label="$t('settings.account.password.update.logoutOtherDevices')"
      color="primary"
      density="compact"
      hide-details
      data-test="settings-account-password-logout-other-devices"
    />
  </OrganismModal>
</template>
