<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { LoginWithPwdRequest } from '~/features/auth/composables/auth.api'
import { object, string } from 'yup'
import { NOT_AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'
import { useForm } from '~/utils/forms/use-form'
import { safeRedirect } from '~/utils/redirect'

definePageMeta({
  layout: 'auth',
  authGroups: [NOT_AUTHENTICATED_ROUTE_GROUP],
})

const route = useRoute()
const { t } = useI18n()
const { login } = useAuth()

const loading = ref<boolean>(false)
const error = ref<string>()

async function onSubmit(values: LoginWithPwdRequest): Promise<void> {
  error.value = undefined
  loading.value = true

  const res = await login(values)
  if (res === 'ok') {
    await navigateTo(safeRedirect(route.query.redirect))

    return
  }

  error.value = t(`login.error.${res}`)

  loading.value = false
}

const { field, handleSubmit } = useForm({
  schema: object({
    username: string().required().label(t('login.username')),
    password: string().required().label(t('login.password')),
  }),
  initialValues: {
    username: '',
    password: '',
  },
  onSubmit,
})
</script>

<template>
  <AuthCard
    :title="$t('login.title')"
    :subtitle="$t('login.subtitle')"
    :error
    :loading="loading"
    :on-submit="handleSubmit"
    :submit-text="$t('login.submit')"
  >
    <VTextField
      v-bind="field('username').props"
      type="text"
      autocomplete="username"
      data-test="login-username"
    />

    <AtomInputPassword
      v-bind="field('password').props"
      data-test="login-password"
    />
  </AuthCard>
</template>
