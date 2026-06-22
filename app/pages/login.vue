<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { LoginRequestDto } from '#shared/dto/identity'
import { object, string } from 'yup'
import { accountNameRule } from '~/features/auth/utils/account-name'
import { useForm } from '~/utils/forms/use-form'

definePageMeta({ layout: 'auth' })

const { t } = useI18n()
const route = useRoute()
const { login, loading } = useAuth()

const formError = ref('')

const notice = computed(() => {
  const reason = route.query.reason
  if (reason === 'expired') {
    return t('login.notice.expired')
  }

  if (reason === 'setupClosed') {
    return t('login.notice.setupClosed')
  }

  return ''
})

function safeRedirect(raw: unknown): string {
  return typeof raw === 'string' && raw.startsWith('/') && !raw.startsWith('//') ? raw : '/'
}

function mapError(status: number): string {
  if (status === 401) {
    return t('login.error.invalid')
  }

  if (status === 429) {
    return t('login.error.tooMany')
  }

  return t('login.error.generic')
}

async function onSubmit(values: LoginRequestDto): Promise<void> {
  formError.value = ''
  const res = await login(values)
  if (res.success) {
    await navigateTo(safeRedirect(route.query.redirect))

    return
  }

  formError.value = mapError(res.error.status)
}

const { field, handleSubmit } = useForm({
  schema: object({
    accountName: accountNameRule(t('login.accountName')),
    password: string().required().label(t('login.password')),
  }),
  initialValues: {
    accountName: '',
    password: '',
  },
  onSubmit,
})
</script>

<template>
  <AuthCard
    :title="$t('login.title')"
    :subtitle="$t('login.subtitle')"
    :error="formError"
    :loading="loading"
    :on-submit="handleSubmit"
    :notice
    :submit-text="$t('login.submit')"
  >
    <VTextField
      v-bind="field('accountName').props"
      type="text"
      autocomplete="username"
      data-test="login-accountName"
    />

    <AtomInputPassword
      v-bind="field('password').props"
      data-test="login-password"
    />
  </AuthCard>
</template>
