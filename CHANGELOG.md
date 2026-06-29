# Changelog

## 1.0.0 (2026-06-29)


### ⚠ BREAKING CHANGES

* **identity:** app_user identifier is now account_name; email removed.

### Features

* account editing — self-service display name & password, admin display name (Lot B) ([#19](https://github.com/Kharente-Deuh/Uchiyomi/issues/19)) ([fa5c6fe](https://github.com/Kharente-Deuh/Uchiyomi/commit/fa5c6fe8ef5278b741dc3ecf1d7f247da5484d72))
* add horizontal logo images ([de4ace6](https://github.com/Kharente-Deuh/Uchiyomi/commit/de4ace6f6a27d1e8da4dd5a46a5d68499b4c33ce))
* add local suwayomi dev environment ([ba45916](https://github.com/Kharente-Deuh/Uchiyomi/commit/ba459160d172e8ee4e936ac8dfb8adaceb1fa971))
* add project base ([27073c1](https://github.com/Kharente-Deuh/Uchiyomi/commit/27073c1798f03e1416687d8dbac946973cd591b8))
* **db:** add cached catalogue models Series and Chapter ([05071c7](https://github.com/Kharente-Deuh/Uchiyomi/commit/05071c76c03c606bf57d48c114ac041c0c84a985))
* **db:** add globally-paced download queue model ([7b53979](https://github.com/Kharente-Deuh/Uchiyomi/commit/7b539798087068663f556cc715566e71067d1e22))
* **db:** add initial overlay migration ([c715e37](https://github.com/Kharente-Deuh/Uchiyomi/commit/c715e37625c52829eaa8190abaa976a961df3407))
* **db:** add overlay enums and extension activation models ([64dee1d](https://github.com/Kharente-Deuh/Uchiyomi/commit/64dee1d36671bcc681a757b076086f4f6d6a98ae))
* **db:** overlay data model — schema, migration & constraints (M1.1) ([08a927f](https://github.com/Kharente-Deuh/Uchiyomi/commit/08a927fcf4b5f7289ddfe7fc3d5f3be7a3ed125b))
* **db:** rework library/progress for cached catalogue and add per-type reading preferences ([d9ce481](https://github.com/Kharente-Deuh/Uchiyomi/commit/d9ce48182cb96cfff56b0a29aaf2f16ba264be72))
* **extensions:** admin-managed extensions, NSFW gate & source config (M4.1a) ([#45](https://github.com/Kharente-Deuh/Uchiyomi/issues/45)) ([ed03641](https://github.com/Kharente-Deuh/Uchiyomi/commit/ed036414fb1a8277bd7bc5341c77b8c162961733))
* **extensions:** browse UI for extensions and sources (M4.1b) ([#48](https://github.com/Kharente-Deuh/Uchiyomi/issues/48)) ([2f40e0d](https://github.com/Kharente-Deuh/Uchiyomi/commit/2f40e0d477b0c7d014b39cbd22cabd455917c96b))
* **front:** app shell & navigation (M3.2) ([#14](https://github.com/Kharente-Deuh/Uchiyomi/issues/14)) ([498baab](https://github.com/Kharente-Deuh/Uchiyomi/commit/498baabfca045173a441beb93b726dee2bc8a703))
* **front:** auth UI & bootstrap (M3.3) ([#16](https://github.com/Kharente-Deuh/Uchiyomi/issues/16)) ([d0eed08](https://github.com/Kharente-Deuh/Uchiyomi/commit/d0eed08d47320839f9abd8e21375f39173e55913))
* **front:** front data-access layer (M3.1) ([#13](https://github.com/Kharente-Deuh/Uchiyomi/issues/13)) ([ddcd5d0](https://github.com/Kharente-Deuh/Uchiyomi/commit/ddcd5d0dc5292a5f00e51348a86873cff112fbe1))
* **front:** vuetify-only forms util with yup-locales i18n ([#15](https://github.com/Kharente-Deuh/Uchiyomi/issues/15)) ([449f6ee](https://github.com/Kharente-Deuh/Uchiyomi/commit/449f6eeec1d06be23e8abc2dad851b9c28ad4697))
* **identity:** local auth spine with revocable sessions (M2.2) ([#9](https://github.com/Kharente-Deuh/Uchiyomi/issues/9)) ([9eb6393](https://github.com/Kharente-Deuh/Uchiyomi/commit/9eb6393c7233026e4264866c0b696084f0d39a34))
* **identity:** replace email login identifier with account name (Lot A) ([#18](https://github.com/Kharente-Deuh/Uchiyomi/issues/18)) ([4fb83d5](https://github.com/Kharente-Deuh/Uchiyomi/commit/4fb83d5f75318abbf42347c85156e0e80c22774f))
* local Suwayomi dev environment (M0.3) ([#3](https://github.com/Kharente-Deuh/Uchiyomi/issues/3)) ([ba45916](https://github.com/Kharente-Deuh/Uchiyomi/commit/ba459160d172e8ee4e936ac8dfb8adaceb1fa971))
* **suwayomi:** typed resilient GraphQL client + catalogue read slice (M2.1) ([#7](https://github.com/Kharente-Deuh/Uchiyomi/issues/7)) ([952d69d](https://github.com/Kharente-Deuh/Uchiyomi/commit/952d69d1d19b58671b3ff874661a249558e36303))
* **suwayomi:** typed resilient graphql client and catalogue read slice ([952d69d](https://github.com/Kharente-Deuh/Uchiyomi/commit/952d69d1d19b58671b3ff874661a249558e36303))
* **ui:** replace MDI with Font Awesome 6 via Iconify (no UnoCSS) ([70776c5](https://github.com/Kharente-Deuh/Uchiyomi/commit/70776c550cfc525c49f9ad249d11a95017049e0c))


### Bug Fixes

* lint ([c9a4782](https://github.com/Kharente-Deuh/Uchiyomi/commit/c9a47824ec937b34644f3f8072ab6a7bf8594271))


### Performance Improvements

* add deps to optimize ([a290ea1](https://github.com/Kharente-Deuh/Uchiyomi/commit/a290ea1acc37e945b43e6f8e66c943e9443eb6a1))
