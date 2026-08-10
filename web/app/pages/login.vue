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
const { providers, fetchProviders } = useOIDCProviders()

const redirect = computed(() => safeRedirect(route.query.redirect))

const oidcErrorCodes = ['oidcState', 'oidcDenied', 'oidcNotAllowed', 'oidcNoAccount', 'oidcUnavailable'] as const
const initialErrorCode = typeof route.query.error === 'string' && (oidcErrorCodes as readonly string[]).includes(route.query.error)
  ? route.query.error
  : undefined

const loading = ref<boolean>(false)
const error = ref<string | undefined>(initialErrorCode ? t(`login.error.${initialErrorCode}`) : undefined)

onMounted(fetchProviders)

async function onSubmit(values: LoginWithPwdRequest): Promise<void> {
  error.value = undefined
  loading.value = true

  const res = await login(values)
  if (res === 'ok') {
    await navigateTo(redirect.value)

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

  <template v-if="providers.length">
    <div class="d-flex ga-4 align-center">
      <VDivider />
      <span class="text-uppercase text-medium-emphasis">{{ $t('common.or') }}</span>
      <VDivider />
    </div>
    <AuthOidcProviderButtons :providers :redirect="redirect" />
  </template>
</template>
