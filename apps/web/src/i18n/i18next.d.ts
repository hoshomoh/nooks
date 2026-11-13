import "i18next"

/*
i18next 27 checks that a call passes every {{variable}} its string has. It can only know
that from typed resources, which nooks does not declare, so with plain string keys the
check has nothing to read and instead makes options required on every call: `t` then
no longer fits the Translate type a helper takes. Off, as the migration guide says.
*/
declare module "i18next" {
  interface CustomTypeOptions {
    strictInterpolationChecks: false
  }
}
