// SPDX-License-Identifier: AGPL-3.0-or-later
// Custom Vuetify icon set backed by Iconify (ADR-0009: no UnoCSS). Font Awesome 6
// collections are bundled offline via `addCollection` here (same module as the
// renderer, so they share one Iconify storage instance), and the set renders the
// SVG synchronously from `getIcon` — which works during SSR, unlike the async
// `<Icon>` component (it renders an empty placeholder until client mount).
import type { IconAliases, IconProps, IconSet } from 'vuetify'
import fa6Regular from '@iconify-json/fa6-regular/icons.json'
import fa6Solid from '@iconify-json/fa6-solid/icons.json'
import { addCollection, getIcon } from '@iconify/vue'
import { h } from 'vue'

addCollection(fa6Solid)
addCollection(fa6Regular)

// Vuetify's built-in components reference these aliases (e.g. `$dropdown`).
// Each maps to a Font Awesome 6 icon (solid unless an outline reads better).
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
  component: (props: IconProps) => {
    const name = typeof props.icon === 'string' ? props.icon : ''
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
