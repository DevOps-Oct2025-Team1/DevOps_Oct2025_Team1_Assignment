package cicd.secrets

deny contains msg if {
  step := input.jobs[_].steps[_]
  contains(lower(step.run), "api_key")
  not contains(step.run, "secrets.")
  msg := "Hardcoded API key detected. Use GitHub Secrets."
}
