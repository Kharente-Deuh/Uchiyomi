<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { object, string, ref as yupRef } from 'yup'
import { DEFAULT_PAGE } from '~/constants'
import { NOT_AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'
import { usernameRule } from '~/utils/forms/rules/username.rule'
import { useForm } from '~/utils/forms/use-form'

definePageMeta({
  layout: 'auth',
  authGroups: [NOT_AUTHENTICATED_ROUTE_GROUP],
})

const { t } = useI18n()
const setupApi = createSetupApi()

type Form = DoSetupBody & { confirmPassword: string }
const formError = ref('')
const loading = ref(false)

async function onSubmit(values: Form): Promise<void> {
  loading.value = true

  formError.value = ''
  const res = await setupApi.doSetup({
    username: values.username,
    password: values.password,
  })

  if (res.success) {
    await navigateTo(DEFAULT_PAGE)

    return
  }

  if (res.error.status === 409) {
    await navigateTo({ path: '/login', query: { reason: 'setupClosed' } })

    return
  }

  formError.value = t('setup.error.generic')

  loading.value = false
}

const { field, handleSubmit } = useForm({
  initialValues: {
    username: '',
    password: '',
    confirmPassword: '',
  },
  schema: object({
    username: usernameRule(t('setup.username')),
    password: string().required().min(10).max(72).label(t('setup.password')),
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
      v-bind="field('username').props"
      type="text"
      autocomplete="username"
      data-test="setup-username"
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
