// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IconAliases, IconProps, IconSet } from 'vuetify'
import fa6Regular from '@iconify-json/fa6-regular/icons.json'
import fa6Solid from '@iconify-json/fa6-solid/icons.json'
import { addCollection, getIcon } from '@iconify/vue'
import { h } from 'vue'

// eslint-disable-next-line unicorn/no-top-level-side-effects
addCollection(fa6Solid)
// eslint-disable-next-line unicorn/no-top-level-side-effects
addCollection(fa6Regular)

// eslint-disable-next-line unicorn/no-top-level-side-effects
addCollection({
  prefix: 'uchi',
  icons: {
    'continuous-vertical': {
      body: '<path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 2v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V2m0 20v-6a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v6"></path>',
      width: 24,
      height: 24,
    },
    'screen-width': {
      body: '<path fill="currentColor" d="M4 20h16v2H4zM4 2h16v2H4zm9 7h3l-4-4-4 4h3v6H8l4 4 4-4h-3z"></path>',
      width: 24,
      height: 24,
    },
    'screen-height': {
      body: '<g transform="matrix(0 1 -1 0 512 0)"><path fill="currentColor" d="M32 64c17.7 0 32 14.3 32 32v320c0 17.7-14.3 32-32 32S0 433.7 0 416V96c0-17.7 14.3-32 32-32m214.6 73.4c12.5 12.5 12.5 32.8 0 45.3L205.3 224h229.5l-41.4-41.4c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0l96 96c12.5 12.5 12.5 32.8 0 45.3l-96 96c-12.5 12.5-32.8 12.5-45.3 0s-12.5-32.8 0-45.3l41.3-41.3H205.2l41.4 41.4c12.5 12.5 12.5 32.8 0 45.3s-32.8 12.5-45.3 0l-96-96c-12.5-12.5-12.5-32.8 0-45.3l96-96c12.5-12.5 32.8-12.5 45.3 0M640 96v320c0 17.7-14.3 32-32 32s-32-14.3-32-32V96c0-17.7 14.3-32 32-32s32 14.3 32 32"/></g>',
      width: 512,
      height: 640,
    },
    'screen-fit': {
      body: '<path fill="currentColor" d="m15 3 2.3 2.3-2.89 2.87 1.42 1.42L18.7 6.7 21 9V3zM3 9l2.3-2.3 2.87 2.89 1.42-1.42L6.7 5.3 9 3H3zm6 12-2.3-2.3 2.89-2.87-1.42-1.42L5.3 17.3 3 15v6zm12-6-2.3 2.3-2.87-2.89-1.42 1.42 2.89 2.87L15 21h6z"></path>',
      width: 24,
      height: 24,
    },
  },
})

export const aliases: IconAliases = {
  collapse: 'fa6-solid:chevron-up',
  complete: 'fa6-solid:check',
  cancel: 'fa6-solid:circle-xmark',
  close: 'fa6-solid:xmark',
  delete: 'fa6-solid:circle-xmark',
  clear: 'fa6-solid:circle-xmark',
  success: 'fa6-solid:circle-check',
  info: 'fa6-solid:circle-info',
  warning: 'fa6-solid:triangle-exclamation',
  error: 'fa6-solid:circle-exclamation',
  prev: 'fa6-solid:chevron-left',
  next: 'fa6-solid:chevron-right',
  checkboxOn: 'fa6-solid:square-check',
  checkboxOff: 'fa6-regular:square',
  checkboxIndeterminate: 'fa6-solid:square-minus',
  delimiter: 'fa6-solid:circle',
  sortAsc: 'fa6-solid:arrow-up',
  sortDesc: 'fa6-solid:arrow-down',
  expand: 'fa6-solid:chevron-down',
  menu: 'fa6-solid:bars',
  subgroup: 'fa6-solid:caret-down',
  dropdown: 'fa6-solid:caret-down',
  radioOn: 'fa6-solid:circle-dot',
  radioOff: 'fa6-regular:circle',
  edit: 'fa6-solid:pen',
  ratingEmpty: 'fa6-regular:star',
  ratingFull: 'fa6-solid:star',
  ratingHalf: 'fa6-solid:star-half-stroke',
  loading: 'fa6-solid:spinner',
  first: 'fa6-solid:angles-left',
  last: 'fa6-solid:angles-right',
  unfold: 'fa6-solid:up-down',
  file: 'fa6-regular:file',
  plus: 'fa6-solid:plus',
  minus: 'fa6-solid:minus',
  calendar: 'fa6-regular:calendar',
  treeviewCollapse: 'fa6-solid:chevron-down',
  treeviewExpand: 'fa6-solid:chevron-right',
  eyeDropper: 'fa6-solid:eye-dropper',
  upload: 'fa6-solid:upload',
  color: 'fa6-solid:palette',
  command: 'fa6-solid:keyboard',
  ctrl: 'fa6-solid:keyboard',
  space: 'fa6-solid:keyboard',
  shift: 'fa6-solid:arrow-up',
  alt: 'fa6-solid:keyboard',
  enter: 'fa6-solid:keyboard',
  arrowup: 'fa6-solid:arrow-up',
  arrowdown: 'fa6-solid:arrow-down',
  arrowleft: 'fa6-solid:arrow-left',
  arrowright: 'fa6-solid:arrow-right',
  backspace: 'fa6-solid:delete-left',
}

export const iconify: IconSet = {
  component: (properties: IconProps) => {
    const name = typeof properties.icon === 'string' ? properties.icon : ''
    const data = name ? getIcon(name) : null
    if (!data) {
      return h('svg', { 'viewBox': '0 0 24 24', 'width': '1em', 'height': '1em', 'aria-hidden': 'true' })
    }

    return h('svg', {
      'xmlns': 'http://www.w3.org/2000/svg',
      'viewBox': `0 0 ${data.width} ${data.height}`,
      'width': '1em',
      'height': '1em',
      'role': 'img',
      'aria-hidden': 'true',
      'innerHTML': data.body,
    })
  },
}
