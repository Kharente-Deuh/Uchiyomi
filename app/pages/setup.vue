<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SetupRequestDto } from '#shared/dto/identity'
import { object, string, ref as yupRef } from 'yup'
import { accountNameRule } from '~/features/auth/utils/account-name'
import { useForm } from '~/utils/forms/use-form'

definePageMeta({ layout: 'auth' })

const { t } = useI18n()
const { setup, loading, minPasswordLength } = useAuth()

const formError = ref('')

type Form = SetupRequestDto & { confirmPassword: string }

async function onSubmit(values: Form): Promise<void> {
  formError.value = ''
  const res = await setup({
    accountName: values.accountName,
    displayName: values.displayName,
    password: values.password,
  })

  if (res.success) {
    await navigateTo('/')

    return
  }

  if (res.error.status === 409) {
    await navigateTo({ path: '/login', query: { reason: 'setupClosed' } })

    return
  }

  formError.value = t('setup.error.generic')
}

const { field, handleSubmit } = useForm({
  initialValues: {
    accountName: '',
    displayName: '',
    password: '',
    confirmPassword: '',
  },
  schema: object({
    accountName: accountNameRule(t('setup.accountName')),
    displayName: string().required().label(t('setup.displayName')),
    password: string().required().min(minPasswordLength.value).label(t('setup.password')),
    confirmPassword: string()
      .required()
      .oneOf([yupRef('password')], t('setup.passwordMismatch'))
      .label(t('setup.confirmPassword')),
  }),
  onSubmit,
})
</script>

<template>
  <AuthCard
    :title="$t('setup.title')"
    :subtitle="$t('setup.subtitle')"
    :error="formError"
    :loading="loading"
    :on-submit="handleSubmit"
    :submit-text="$t('setup.submit')"
  >
    <VTextField
      v-bind="field('accountName').props"
      type="text"
      autocomplete="username"
      data-test="setup-accountName"
    />
    <VTextField
      v-bind="field('displayName').props"
      data-test="setup-displayName"
    />
    <AtomInputPassword
      v-bind="field('password').props"
      data-test="setup-password"
    />
    <AtomInputPassword
      v-bind="field('confirmPassword').props"
      data-test="setup-confirm"
    />
  </AuthCard>
</template>
