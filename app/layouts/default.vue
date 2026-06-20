<template>
  <VApp>
    <VAppBar flat density="comfortable">
      <VAppBarTitle>{{ $t('app.name') }}</VAppBarTitle>
      <template #append>
        <LanguageSwitch />
        <DarkToggle />
      </template>
    </VAppBar>
    <VMain>
      <VContainer>
        <slot />
      </VContainer>
    </VMain>
  </VApp>
</template>
