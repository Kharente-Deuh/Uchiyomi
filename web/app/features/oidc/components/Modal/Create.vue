<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { array, boolean, object, string } from 'yup'
import { useForm } from '~/utils/forms/use-form'

const show = defineModel<boolean>({ required: true })

const { t } = useI18n()
const { loading, create } = useOidc()

async function onSubmit(values: CreateOidcProviderRequest): Promise<void> {
  await create(values)

  show.value = false
}

const { field, handleSubmit, reset, isValid } = useForm({
  schema: object({
    displayName: string().required().label(t('settings.oidc.fields.displayName.label')),
    issuerUrl: string().required().url().label(t('settings.oidc.fields.issuerUrl.label')),
    clientId: string().required().label(t('settings.oidc.fields.clienId.label')),
    clientSecret: string().required().label(t('settings.oidc.fields.clientSecret.label')),
    autoProvision: boolean().required().label(t('settings.oidc.fields.autoProvision.label')),
    scopes: array().of(string().required()).required().min(1).label(t('settings.oidc.fields.scopes.label')),
    usernameClaim: string ().required().label(t('settings.oidc.fields.usernameClame.label')),
    roleClaim: string().defined().label(t('settings.oidc.fields.roleClaim.label')),
    adminValues: array().of(string().required()).defined().label(t('settings.oidc.fields.adminValues.label')),
    allowedValues: array().of(string().required()).defined().label(t('settings.oidc.fields.allowedValues.label')),
  }),
  initialValues: {
    displayName: '',
    issuerUrl: '',
    clientId: '',
    clientSecret: '',
    autoProvision: false,
    scopes: [],
    usernameClaim: '',
    roleClaim: '',
    adminValues: [],
    allowedValues: [],
  },
  onSubmit,
})

const autoProvisionProps = computed(() => {
  const { 'onUpdate:modelValue': onUpdate, ...props } = field('autoProvision').props

  return { ...props, 'onUpdate:modelValue': (value: boolean | null) => onUpdate(value ?? false) }
})

watch(show, (val: boolean): void => {
  if (val) {
    reset()
  }
})

watch(() => field('roleClaim').props.modelValue, (value: string) => {
  if (!value) {
    field('allowedValues').props['onUpdate:modelValue']([])
    field('adminValues').props['onUpdate:modelValue']([])
  }
})
</script>

<template>
  <OrganismModal
    v-model="show"
    :title="$t('settings.oidc.create.title')"
    prepend-icon="fa6-solid:plus"
    :is-form-complete="isValid"
    :loading
    :submit-text="$t('actions.create')"
    @submit="handleSubmit"
    @cancel="show = false"
  >
    <div class="d-flex ga-4 text-truncate w-100 align-center pb-2">
      <VIcon icon="fa6-regular:address-card" color="secondary" />

      <span class="text-title-large font-title text-truncate">{{ $t('settings.oidc.category.providerInfos.title') }}</span>
    </div>
    <VTextField v-bind="field('displayName').props" />
    <VTextField v-bind="field('issuerUrl').props">
      <template #append>
        <OidcBtnTest :issuer-url="field('issuerUrl').props.modelValue" />
      </template>
    </VTextField>
    <VTextField v-bind="field('clientId').props" />
    <AtomInputPassword v-bind="field('clientSecret').props" />
    <VCheckbox
      density="compact"
      :messages="[$t('settings.oidc.fields.autoProvision.hint')]"
      v-bind="autoProvisionProps"
    />

    <div class="d-flex ga-4 text-truncate w-100 align-center py-2">
      <VIcon icon="fa6-solid:key" color="secondary" />

      <span class="text-title-large font-title text-truncate">{{ $t('settings.oidc.category.claims.title') }}</span>
    </div>

    <VCombobox
      v-bind="field('scopes').props"
      clear-icon="fa6-solid:xmark"
      density="comfortable"
      bg-color="surface-variant"
      class="text-field-override"
      color="primary"
      variant="outlined"
      menu-icon="fa6-solid:caret-down"
      chips
      clearable
      multiple
    />
    <VTextField v-bind="field('usernameClaim').props" />
    <VTextField v-bind="field('roleClaim').props" clearable />
    <VCombobox
      v-bind="field('adminValues').props"
      clear-icon="fa6-solid:xmark"
      density="comfortable"
      bg-color="surface-variant"
      class="text-field-override"
      :disabled="!field('roleClaim').props.modelValue"
      color="primary"
      variant="outlined"
      menu-icon="fa6-solid:caret-down"
      chips
      clearable
      multiple
    />
    <VCombobox
      v-bind="field('allowedValues').props"
      clear-icon="fa6-solid:xmark"
      density="comfortable"
      :disabled="!field('roleClaim').props.modelValue"
      bg-color="surface-variant"
      class="text-field-override"
      color="primary"
      variant="outlined"
      menu-icon="fa6-solid:caret-down"
      chips
      clearable
      multiple
    />
  </OrganismModal>
</template>
