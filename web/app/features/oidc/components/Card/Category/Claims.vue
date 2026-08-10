<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { array, object, string } from 'yup'
import { useForm } from '~/utils/forms/use-form'

defineProps<{
  loading: boolean
}>()

const { t } = useI18n()
const { provider, update, updateLoading } = useOidcProvider()

type Form = Pick<UpdateOidcProviderRequest, 'scopes' | 'usernameClaim' | 'roleClaim' | 'adminValues' | 'allowedValues'>

async function onSubmit(values: Form): Promise<void> {
  if (!provider.value) {
    return
  }

  const { users: _, ...data } = provider.value

  await update({ ...data, ...values })
}

const { field, handleSubmit, reset, isValid } = useForm({
  schema: object({
    scopes: array().of(string().required()).required().min(1).label(t('settings.oidc.fields.scopes.label')),
    usernameClaim: string().required().label(t('settings.oidc.fields.usernameClame.label')),
    roleClaim: string().defined().label(t('settings.oidc.fields.roleClaim.label')),
    adminValues: array().of(string().required()).defined().label(t('settings.oidc.fields.adminValues.label')),
    allowedValues: array().of(string().required()).defined().label(t('settings.oidc.fields.allowedValues.label')),
  }),
  initialValues: {
    scopes: [],
    usernameClaim: '',
    roleClaim: '',
    adminValues: [],
    allowedValues: [],
  },
  onSubmit,
})

watch(provider, (value?: OidcProvider) => {
  if (value) {
    reset({
      scopes: value.scopes,
      usernameClaim: value.usernameClaim,
      roleClaim: value.roleClaim ?? '',
      adminValues: value.adminValues ?? [],
      allowedValues: value.allowedValues ?? [],
    })
  }
}, { immediate: true })

const hasChanged = computed(() => {
  if (!provider.value) {
    return false
  }

  return !areArrayEquals(provider.value.scopes, field('scopes').props.modelValue)
    || provider.value.usernameClaim !== field('usernameClaim').props.modelValue
    || (provider.value.roleClaim || '') !== field('roleClaim').props.modelValue
    || !areArrayEquals(provider.value.adminValues ?? [], field('adminValues').props.modelValue)
    || !areArrayEquals(provider.value.allowedValues ?? [], field('allowedValues').props.modelValue)
})

function areArrayEquals(a: string[], b: string[]): boolean {
  if (a.length !== b.length) {
    return false
  }

  return a.every(item => b.includes(item))
}

watch(() => field('roleClaim').props.modelValue, (value: string) => {
  if (!value) {
    field('allowedValues').props['onUpdate:modelValue']([])
    field('adminValues').props['onUpdate:modelValue']([])
  }
})

const isFormValid = computed(() => isValid.value && hasChanged.value)
</script>

<template>
  <MoleculeCardComponent
    :disabled="loading || updateLoading"
    :loading
    :title="$t('settings.oidc.category.claims.title')"
    icon="fa6-solid:key"
  >
    <div class="provider-claims-grid">
      <VCombobox v-bind="field('scopes').props" />
      <VTextField v-bind="field('usernameClaim').props" />
      <VTextField
        v-bind="field('roleClaim').props"
        clearable
        @update:model-value="field('roleClaim').props['onUpdate:modelValue']($event ?? '')"
      />
      <VCombobox
        v-bind="field('adminValues').props"
        :disabled="!field('roleClaim').props.modelValue"
      />
      <VCombobox
        v-bind="field('allowedValues').props"
        :disabled="!field('roleClaim').props.modelValue"
      />
    </div>

    <div class="w-100 mt-6 d-flex justify-end">
      <VBtn
        :text="$t('actions.save')"
        :loading="updateLoading"
        color="secondary"
        class="border-thin-secondary"
        :disabled="!loading && !isFormValid"
        @click="handleSubmit"
      />
    </div>
  </MoleculeCardComponent>
</template>

<style lang="scss">
.provider-claims-grid {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(2, 1fr);
}

@media (max-width: 800px) {
  .provider-claims-grid {
    grid-template-columns: 1fr;
  }
}
</style>
