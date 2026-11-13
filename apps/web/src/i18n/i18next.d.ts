import "i18next"

// i18next 27 checks every {{variable}} is passed, which needs typed resources. nooks has
// none, so the check only made options required on every call. Off, as its guide says.
declare module "i18next" {
  interface CustomTypeOptions {
    strictInterpolationChecks: false
  }
}
